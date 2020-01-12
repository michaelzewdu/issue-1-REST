
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

--
-- Name: issue#1; Type: SCHEMA; Schema: -; Owner: issue#1_dev
--

DROP SCHEMA "public";
CREATE SCHEMA "issue#1";

ALTER DATABASE "issue#1_db" SET search_path = "issue#1";

ALTER SCHEMA "issue#1" OWNER TO "issue#1_dev";
