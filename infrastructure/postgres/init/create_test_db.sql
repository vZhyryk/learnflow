-- Runs once, only on first init of the sonar_db data volume (official Postgres
-- image convention: /docker-entrypoint-initdb.d/*.sql executes only against an
-- empty PGDATA). Creates a database dedicated to integration tests, isolated
-- from the 'sonar' database SonarQube owns in the same instance.
--
-- SUPERUSER: migration 000001 runs `CREATE EXTENSION pg_stat_statements`,
-- which is not a trusted extension — only a superuser can create it. This
-- role only ever exists inside a throwaway CI/local test container, so
-- granting superuser here carries none of the risk it would in a real
-- environment.
CREATE USER learnflow_test WITH SUPERUSER PASSWORD 'learnflow_test';
CREATE DATABASE learnflow_test OWNER learnflow_test;