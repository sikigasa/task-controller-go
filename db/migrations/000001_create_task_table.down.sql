DROP TRIGGER IF EXISTS set_updated_at ON "task";
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS "task_tag";
DROP TABLE IF EXISTS "tag";
DROP TABLE IF EXISTS "task";