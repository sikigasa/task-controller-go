package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sikigasa/task-controller/internal/infra"
	postgresDriver "github.com/sikigasa/task-controller/internal/infra/driver"
	task "github.com/sikigasa/task-controller/proto/v1"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	// PostgreSQLコンテナの起動
	postgresContainer, err := postgres.Run(ctx,
		"postgres:17.5-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	// データベース接続文字列の取得
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// データベース接続
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// 接続確認
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// テーブル作成
	if err := createTables(db); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// クリーンアップ関数
	cleanup := func() {
		db.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS task (
            id VARCHAR(255) PRIMARY KEY,
            title VARCHAR(255) NOT NULL,
            description TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            limited_at TIMESTAMP,
            is_end BOOLEAN DEFAULT FALSE
        )`,
		`CREATE TABLE IF NOT EXISTS tag (
            id VARCHAR(255) PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE
        )`,
		`CREATE TABLE IF NOT EXISTS task_tag (
            task_id VARCHAR(255) REFERENCES task(id) ON DELETE CASCADE,
            tag_id VARCHAR(255) REFERENCES tag(id) ON DELETE CASCADE,
            PRIMARY KEY (task_id, tag_id)
        )`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func setupTestService(t *testing.T, db *sql.DB) task.TaskServiceServer {
	taskRepo := infra.NewTaskRepo(db)
	tagRepo := infra.NewTagRepo(db)
	taskTagRepo := infra.NewTaskTagRepo(db)
	tx := postgresDriver.NewPostgresTransaction(db)

	return NewTaskService(taskRepo, tagRepo, taskTagRepo, tx)
}

func createTestTag(t *testing.T, db *sql.DB, id, name string) {
	query := `INSERT INTO tag (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`
	_, err := db.Exec(query, id, name)
	if err != nil {
		t.Fatalf("failed to create test tag: %v", err)
	}
}

func TestTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	taskService := setupTestService(t, db)

	t.Run("CreateTask", func(t *testing.T) {
		testCreateTask(t, taskService, db)
	})

	t.Run("GetTask", func(t *testing.T) {
		testGetTask(t, taskService, db)
	})

	t.Run("ListTask", func(t *testing.T) {
		testListTask(t, taskService, db)
	})

	t.Run("UpdateTask", func(t *testing.T) {
		testUpdateTask(t, taskService, db)
	})

	t.Run("DeleteTask", func(t *testing.T) {
		testDeleteTask(t, taskService, db)
	})
}

func testCreateTask(t *testing.T, taskService task.TaskServiceServer, db *sql.DB) {
	t.Run("正常系_タグありの場合", func(t *testing.T) {
		// テスト用タグを作成
		createTestTag(t, db, "tag1", "テストタグ1")
		createTestTag(t, db, "tag2", "テストタグ2")

		req := &task.CreateTaskRequest{
			Title:       "テストタスク",
			Description: "テストの説明",
			LimitedAt:   timestamppb.New(time.Now().Add(24 * time.Hour)),
			TagIds:      []string{"tag1", "tag2"},
		}

		res, err := taskService.CreateTask(context.Background(), req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if res == nil {
			t.Errorf("expected response, got nil")
			return
		}

		if res.Id == "" {
			t.Errorf("expected non-empty ID, got empty string")
		}

		// UUIDの形式チェック
		if _, err := uuid.Parse(res.Id); err != nil {
			t.Errorf("expected valid UUID, got %v", res.Id)
		}
	})

	t.Run("正常系_タグなしの場合", func(t *testing.T) {
		req := &task.CreateTaskRequest{
			Title:       "タグなしタスク",
			Description: "タグなしの説明",
			LimitedAt:   timestamppb.New(time.Now().Add(24 * time.Hour)),
			TagIds:      []string{},
		}

		res, err := taskService.CreateTask(context.Background(), req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if res == nil || res.Id == "" {
			t.Errorf("expected valid response with ID")
		}
	})
}

func testGetTask(t *testing.T, taskService task.TaskServiceServer, db *sql.DB) {
	t.Run("正常系", func(t *testing.T) {
		// テスト用タグとタスクを作成
		createTestTag(t, db, "get_tag1", "取得テストタグ")

		createReq := &task.CreateTaskRequest{
			Title:       "取得テストタスク",
			Description: "取得テストの説明",
			LimitedAt:   timestamppb.New(time.Now().Add(24 * time.Hour)),
			TagIds:      []string{"get_tag1"},
		}

		createRes, err := taskService.CreateTask(context.Background(), createReq)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		// 作成したタスクを取得
		getReq := &task.GetTaskRequest{
			Id: createRes.Id,
		}

		res, err := taskService.GetTask(context.Background(), getReq)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if res == nil || res.Task == nil {
			t.Errorf("expected task response, got nil")
		}

		if res.Task.Title != "取得テストタスク" {
			t.Errorf("expected title '取得テストタスク', got %v", res.Task.Title)
		}

		if len(res.Task.Tags) != 1 {
			t.Errorf("expected 1 tag, got %d", len(res.Task.Tags))
		}
	})

	t.Run("異常系_存在しないタスク", func(t *testing.T) {
		getReq := &task.GetTaskRequest{
			Id: "non-existent-id",
		}

		_, err := taskService.GetTask(context.Background(), getReq)
		if err == nil {
			t.Errorf("expected error for non-existent task, got nil")
		}
	})
}

func testListTask(t *testing.T, taskService task.TaskServiceServer, db *sql.DB) {
	t.Run("正常系", func(t *testing.T) {
		// 複数のテストタスクを作成
		for i := 0; i < 3; i++ {
			req := &task.CreateTaskRequest{
				Title:       "リストテストタスク",
				Description: "リストテストの説明",
				LimitedAt:   timestamppb.New(time.Now().Add(24 * time.Hour)),
				TagIds:      []string{},
			}
			_, err := taskService.CreateTask(context.Background(), req)
			if err != nil {
				t.Fatalf("failed to create task %d: %v", i, err)
			}
		}

		req := &task.ListTaskRequest{
			Limit:  10,
			Offset: 0,
		}

		res, err := taskService.ListTask(context.Background(), req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if res == nil {
			t.Errorf("expected response, got nil")
			return
		}

		if len(res.Tasks) == 0 {
			t.Errorf("expected tasks, got empty list")
		}
	})

	t.Run("正常系_ページネーション", func(t *testing.T) {
		req := &task.ListTaskRequest{
			Limit:  2,
			Offset: 0,
		}

		res, err := taskService.ListTask(context.Background(), req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if res == nil {
			t.Errorf("expected response, got nil")
		}

		if len(res.Tasks) > 2 {
			t.Errorf("expected at most 2 tasks, got %d", len(res.Tasks))
		}
	})
}

func testUpdateTask(t *testing.T, taskService task.TaskServiceServer, db *sql.DB) {
	t.Run("正常系", func(t *testing.T) {
		// テスト用タスクを作成
		createReq := &task.CreateTaskRequest{
			Title:       "更新前タスク",
			Description: "更新前の説明",
			LimitedAt:   timestamppb.New(time.Now().Add(24 * time.Hour)),
			TagIds:      []string{},
		}

		createRes, err := taskService.CreateTask(context.Background(), createReq)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		// タスクを更新
		updateReq := &task.UpdateTaskRequest{
			Id:          createRes.Id,
			Title:       "更新後タスク",
			Description: "更新後の説明",
			LimitedAt:   timestamppb.New(time.Now().Add(48 * time.Hour)),
			IsEnd:       true,
			TagIds:      []string{},
		}

		res, err := taskService.UpdateTask(context.Background(), updateReq)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if res == nil || !res.Success {
			t.Errorf("expected successful update")
		}

		// 更新されたことを確認
		getReq := &task.GetTaskRequest{
			Id: createRes.Id,
		}

		getRes, err := taskService.GetTask(context.Background(), getReq)
		if err != nil {
			t.Errorf("failed to get updated task: %v", err)
		}

		if getRes.Task.Title != "更新後タスク" {
			t.Errorf("expected updated title '更新後タスク', got %v", getRes.Task.Title)
		}

		if !getRes.Task.IsEnd {
			t.Errorf("expected task to be marked as end")
		}
	})
}

func testDeleteTask(t *testing.T, taskService task.TaskServiceServer, db *sql.DB) {
	t.Run("正常系", func(t *testing.T) {
		// テスト用タスクを作成
		createReq := &task.CreateTaskRequest{
			Title:       "削除テストタスク",
			Description: "削除テストの説明",
			LimitedAt:   timestamppb.New(time.Now().Add(24 * time.Hour)),
			TagIds:      []string{},
		}

		createRes, err := taskService.CreateTask(context.Background(), createReq)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		// タスクを削除
		deleteReq := &task.DeleteTaskRequest{
			Id: createRes.Id,
		}

		res, err := taskService.DeleteTask(context.Background(), deleteReq)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if res == nil || !res.Success {
			t.Errorf("expected successful deletion")
		}

		// 削除されたことを確認
		getReq := &task.GetTaskRequest{
			Id: createRes.Id,
		}

		_, err = taskService.GetTask(context.Background(), getReq)
		if err == nil {
			t.Errorf("expected error when getting deleted task, got nil")
		}
	})

	t.Run("異常系_存在しないタスク", func(t *testing.T) {
		deleteReq := &task.DeleteTaskRequest{
			Id: "non-existent-id",
		}

		_, err := taskService.DeleteTask(context.Background(), deleteReq)
		if err == nil {
			t.Errorf("expected error when deleting non-existent task, got nil")
		}
	})
}
