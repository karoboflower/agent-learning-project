-- Enterprise Agent Platform - Database Initialization Script

-- 创建各服务的数据库
CREATE DATABASE IF NOT EXISTS agent_db;
CREATE DATABASE IF NOT EXISTS task_db;
CREATE DATABASE IF NOT EXISTS tool_db;
CREATE DATABASE IF NOT EXISTS user_db;
CREATE DATABASE IF NOT EXISTS tenant_db;
CREATE DATABASE IF NOT EXISTS cost_db;

-- 为各数据库创建扩展
\c agent_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

\c task_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

\c tool_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c user_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c tenant_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c cost_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "timescaledb";

-- 授予权限
GRANT ALL PRIVILEGES ON DATABASE agent_db TO agent;
GRANT ALL PRIVILEGES ON DATABASE task_db TO agent;
GRANT ALL PRIVILEGES ON DATABASE tool_db TO agent;
GRANT ALL PRIVILEGES ON DATABASE user_db TO agent;
GRANT ALL PRIVILEGES ON DATABASE tenant_db TO agent;
GRANT ALL PRIVILEGES ON DATABASE cost_db TO agent;

\echo 'Database initialization completed!'
