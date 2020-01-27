SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


CREATE ROLE "issue#1_REST";
ALTER ROLE "issue#1_REST" WITH NOSUPERUSER INHERIT NOCREATEROLE NOCREATEDB LOGIN NOREPLICATION NOBYPASSRLS PASSWORD 'md5bea35b6c3bc26f3c5408c166b3b39530';
CREATE ROLE "issue#1_dev";
ALTER ROLE "issue#1_dev" WITH NOSUPERUSER INHERIT CREATEROLE CREATEDB LOGIN NOREPLICATION NOBYPASSRLS PASSWORD 'md5893dd1423078d61f657329366c35f8a3';

DROP DATABASE "issue#1_db";

CREATE DATABASE "issue#1_db" WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'English_United States.1252' LC_CTYPE = 'English_United States.1252';


ALTER DATABASE "issue#1_db" OWNER TO "issue#1_dev";
