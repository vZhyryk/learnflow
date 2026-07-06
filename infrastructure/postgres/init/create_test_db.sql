-- Runs once, only on first init of the sonar_db data volume (official Postgres
-- image convention: /docker-entrypoint-initdb.d/*.sql executes only against an
-- empty PGDATA). Creates a database dedicated to integration tests, isolated
-- from the 'sonar' database SonarQube owns in the same instance.
CREATE USER learnflow_test WITH PASSWORD 'learnflow_test';
CREATE DATABASE learnflow_test OWNER learnflow_test;