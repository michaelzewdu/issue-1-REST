\ir setup-db.sql;
\connect issue#1_db postgres
\ir setup-schema.sql;
\ir setup-tables.sql;
\c issue#1_db issue#1_dev
\dt
