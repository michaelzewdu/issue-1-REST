
--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA "issue#1";


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: setup_user(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".setup_user() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    insert into channels(username, name)
    values (new.username, new.username || '''s channel');
    insert into channel_admins (channel_username, username, is_owner)
    values (new.username, new.username, true);
    insert into feeds(owner_username, sorting)
    values (new.username, 'hot');
    return new;
END;
$$;


ALTER FUNCTION "issue#1".setup_user() OWNER TO "issue#1_dev";

--
-- Name: tsv_comment_trigger(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".tsv_comment_trigger() RETURNS trigger
    LANGUAGE plpgsql
AS $$
DECLARE
    tsv tsvector := ''::tsvector;
BEGIN
    tsv := (SELECT setweight(to_tsvector('english', COALESCE(new.content, '')), 'D')
                       ||
                   setweight(to_tsvector('english', COALESCE(new.commented_by, '')), 'B')
    );
    insert into tsvs_comment (comment_id, vector)
    values (new.id, tsv)
    ON CONFLICT (comment_id)
        do update
        set vector= tsv;
    return new;
END;
$$;


ALTER FUNCTION "issue#1".tsv_comment_trigger() OWNER TO "issue#1_dev";

--
-- Name: tsv_on_metadata_update_trigger(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".tsv_on_metadata_update_trigger() RETURNS trigger
    LANGUAGE plpgsql
AS $$
DECLARE
    tsv tsvector := ''::tsvector;
BEGIN
    case (select r.type from releases as r where id = new.release_id)
        when 'image' then tsv := (SELECT setweight(to_tsvector('english', COALESCE(owner_channel, '')), 'B')
                                             ||
                                         setweight(to_tsvector('english', 'image'), 'A')
                                             ||
                                         setweight(to_tsvector('english', COALESCE(new.description, '')), 'D')
                                             ||
                                         setweight(to_tsvector('english', new.other), 'C')
                                             ||
                                         setweight(to_tsvector('english', COALESCE(new.genre_defining, '')), 'B')
                                             ||
                                         setweight(to_tsvector('english', COALESCE(new.title, '')), 'A')
                                  FROM (SELECT owner_channel
                                        FROM releases
                                        WHERE id = new.release_id) as r2
        );
                          insert into tsvs_release (release_id, vector)
                          values (new.release_id, tsv)
                          ON CONFLICT (release_id) do update
                              set vector= tsv;
        when 'text' then tsv := (SELECT setweight(to_tsvector('english', COALESCE(owner_channel, '')), 'B')
                                            ||
                                        setweight(to_tsvector('english', 'text'), 'A')
                                            ||
                                        setweight(to_tsvector('english', COALESCE(new.description, '')), 'D')
                                            ||
                                        setweight(to_tsvector('english', new.other), 'C')
                                            ||
                                        setweight(to_tsvector('english', COALESCE(new.genre_defining, '')), 'B')
                                            ||
                                        setweight(to_tsvector('english', COALESCE(new.title, '')), 'A')
                                            ||
                                        setweight(to_tsvector('english', COALESCE(content, '')), 'D')
                                 FROM (SELECT id as release_id, owner_channel
                                       FROM releases
                                       WHERE id = new.release_id
                                      ) as r2
                                          left join releases_text_based on releases_text_based.release_id = r2.release_id
        );
                         insert into tsvs_release (release_id, vector)
                         values (new.release_id, tsv)
                         ON CONFLICT (release_id) do update
                             set vector= tsv;
        else
        end case;
    return new;
END;
$$;


ALTER FUNCTION "issue#1".tsv_on_metadata_update_trigger() OWNER TO "issue#1_dev";

--
-- Name: tsv_post_trigger(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".tsv_post_trigger() RETURNS trigger
    LANGUAGE plpgsql
AS $$
DECLARE
    tsv tsvector := ''::tsvector;
BEGIN
    tsv := (SELECT setweight(to_tsvector('english', COALESCE(new.description, '')), 'D')
                       ||
                   setweight(to_tsvector('english', COALESCE(new.title, '')), 'A')
                       ||
                   setweight(to_tsvector('english', COALESCE(new.posted_by, '')), 'B')
                       ||
                   setweight(to_tsvector('english', COALESCE(new.channel_from, '')), 'B')
    );
    insert into tsvs_posts (post_id, vector)
    values (new.id, tsv)
    ON CONFLICT (post_id)
        do update
        set vector= tsv;
    return new;
END;
$$;


ALTER FUNCTION "issue#1".tsv_post_trigger() OWNER TO "issue#1_dev";

--
-- Name: tsv_text_based_update_trigger(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".tsv_text_based_update_trigger() RETURNS trigger
    LANGUAGE plpgsql
AS $$
DECLARE
    tsv tsvector := ''::tsvector;
BEGIN
    tsv := (SELECT setweight(to_tsvector('english', COALESCE(owner_channel, '')), 'B')
                       ||
                   setweight(to_tsvector('english', 'text'), 'A')
                       ||
                   setweight(to_tsvector('english', COALESCE(description, '')), 'D')
                       ||
                   setweight(to_tsvector('english', other), 'C')
                       ||
                   setweight(to_tsvector('english', COALESCE(genre_defining, '')), 'B')
                       ||
                   setweight(to_tsvector('english', COALESCE(title, '')), 'A')
                       ||
                   setweight(to_tsvector('english', COALESCE(new.content, '')), 'D')
            FROM (SELECT id as release_id, owner_channel
                  FROM releases
                  WHERE id = new.release_id
                 ) as r2
                     left join release_metadata on release_metadata.release_id = r2.release_id
    );
    insert into tsvs_release (release_id, vector)
    values (new.release_id, tsv)
    ON CONFLICT (release_id) do update
        set vector= tsv;
    return new;
END;
$$;


ALTER FUNCTION "issue#1".tsv_text_based_update_trigger() OWNER TO "issue#1_dev";

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: channel_admins; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_admins (
                                          channel_username character varying(24) NOT NULL,
                                          username character varying(24) NOT NULL,
                                          is_owner boolean NOT NULL
);


ALTER TABLE "issue#1".channel_admins OWNER TO "issue#1_dev";

--
-- Name: channel_official_catalog; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_official_catalog (
                                                    channel_username character varying(24) NOT NULL,
                                                    release_id integer NOT NULL,
                                                    post_from_id integer NOT NULL
);


ALTER TABLE "issue#1".channel_official_catalog OWNER TO "issue#1_dev";

--
-- Name: channel_pictures; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_pictures (
                                            channelname character varying(24) NOT NULL,
                                            image_name text NOT NULL
);


ALTER TABLE "issue#1".channel_pictures OWNER TO "issue#1_dev";

--
-- Name: channel_stickies; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_stickies (
    post_id integer
);


ALTER TABLE "issue#1".channel_stickies OWNER TO "issue#1_dev";

--
-- Name: channels; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channels (
                                    username character varying(24) NOT NULL,
                                    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
                                    name character varying(80) NOT NULL,
                                    description text
);


ALTER TABLE "issue#1".channels OWNER TO "issue#1_dev";

--
-- Name: comments; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".comments (
                                    post_from integer NOT NULL,
                                    id integer NOT NULL,
                                    reply_to integer NOT NULL,
                                    content text NOT NULL,
                                    commented_by character varying(24) NOT NULL,
                                    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "issue#1".comments OWNER TO "issue#1_dev";

--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE "issue#1".comments ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME "issue#1".comments_id_seq
        START WITH 1
        INCREMENT BY 1
        NO MINVALUE
        NO MAXVALUE
        CACHE 1
    );


--
-- Name: feed_subscriptions; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".feed_subscriptions (
                                              feed_id integer NOT NULL,
                                              channel_username character varying(24) NOT NULL,
                                              subscription_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "issue#1".feed_subscriptions OWNER TO "issue#1_dev";

--
-- Name: feeds; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".feeds (
                                 owner_username character varying(24) NOT NULL,
                                 sorting text NOT NULL,
                                 id integer NOT NULL
);


ALTER TABLE "issue#1".feeds OWNER TO "issue#1_dev";

--
-- Name: feeds_id_seq; Type: SEQUENCE; Schema: issue#1; Owner: issue#1_dev
--

CREATE SEQUENCE "issue#1".feeds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "issue#1".feeds_id_seq OWNER TO "issue#1_dev";

--
-- Name: feeds_id_seq; Type: SEQUENCE OWNED BY; Schema: issue#1; Owner: issue#1_dev
--

ALTER SEQUENCE "issue#1".feeds_id_seq OWNED BY "issue#1".feeds.id;


--
-- Name: post_contents; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".post_contents (
                                         post_id integer NOT NULL,
                                         release_id integer NOT NULL
);


ALTER TABLE "issue#1".post_contents OWNER TO "issue#1_dev";

--
-- Name: posts; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".posts (
                                 id integer NOT NULL,
                                 description text,
                                 title character varying(256) NOT NULL,
                                 posted_by character varying(22) NOT NULL,
                                 channel_from character varying(22) NOT NULL,
                                 creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "issue#1".posts OWNER TO "issue#1_dev";

--
-- Name: post_id_seq; Type: SEQUENCE; Schema: issue#1; Owner: issue#1_dev
--

CREATE SEQUENCE "issue#1".post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "issue#1".post_id_seq OWNER TO "issue#1_dev";

--
-- Name: post_id_seq; Type: SEQUENCE OWNED BY; Schema: issue#1; Owner: issue#1_dev
--

ALTER SEQUENCE "issue#1".post_id_seq OWNED BY "issue#1".posts.id;


--
-- Name: post_stars; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".post_stars (
                                      star_count integer NOT NULL,
                                      post_id integer NOT NULL,
                                      username character varying(22) NOT NULL
);


ALTER TABLE "issue#1".post_stars OWNER TO "issue#1_dev";

--
-- Name: release_metadata; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".release_metadata (
                                            release_id integer NOT NULL,
                                            description text,
                                            other jsonb,
                                            genre_defining text,
                                            release_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
                                            title text
);


ALTER TABLE "issue#1".release_metadata OWNER TO "issue#1_dev";

--
-- Name: releases; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".releases (
                                    id integer NOT NULL,
                                    owner_channel character varying(24) NOT NULL,
                                    type text NOT NULL,
                                    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "issue#1".releases OWNER TO "issue#1_dev";

--
-- Name: releases_image_based; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".releases_image_based (
                                                release_id integer NOT NULL,
                                                image_name text NOT NULL
);


ALTER TABLE "issue#1".releases_image_based OWNER TO "issue#1_dev";

--
-- Name: releases_text_based; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".releases_text_based (
                                               release_id integer NOT NULL,
                                               content text NOT NULL
);


ALTER TABLE "issue#1".releases_text_based OWNER TO "issue#1_dev";

--
-- Name: title_id_seq; Type: SEQUENCE; Schema: issue#1; Owner: issue#1_dev
--

CREATE SEQUENCE "issue#1".title_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "issue#1".title_id_seq OWNER TO "issue#1_dev";

--
-- Name: title_id_seq; Type: SEQUENCE OWNED BY; Schema: issue#1; Owner: issue#1_dev
--

ALTER SEQUENCE "issue#1".title_id_seq OWNED BY "issue#1".releases.id;


--
-- Name: tsvs_comment; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".tsvs_comment (
                                        comment_id integer NOT NULL,
                                        vector tsvector
);


ALTER TABLE "issue#1".tsvs_comment OWNER TO "issue#1_dev";

--
-- Name: tsvs_posts; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".tsvs_posts (
                                      post_id integer NOT NULL,
                                      vector tsvector
);


ALTER TABLE "issue#1".tsvs_posts OWNER TO "issue#1_dev";

--
-- Name: tsvs_release; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".tsvs_release (
                                        release_id integer NOT NULL,
                                        vector tsvector
);


ALTER TABLE "issue#1".tsvs_release OWNER TO "issue#1_dev";

--
-- Name: user_avatars; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".user_avatars (
                                        username character varying(24) NOT NULL,
                                        image_name text NOT NULL
);


ALTER TABLE "issue#1".user_avatars OWNER TO "issue#1_dev";

--
-- Name: user_bookmarks; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".user_bookmarks (
                                          username character varying(24) NOT NULL,
                                          post_id integer NOT NULL,
                                          creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "issue#1".user_bookmarks OWNER TO "issue#1_dev";

--
-- Name: users; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".users (
                                 email "issue#1".citext NOT NULL,
                                 username character varying(24) NOT NULL,
                                 creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
                                 pass_hash text NOT NULL,
                                 first_name character varying(30),
                                 middle_name character varying(30),
                                 last_name character varying(30)
);


ALTER TABLE "issue#1".users OWNER TO "issue#1_dev";

--
-- Name: users_bio; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".users_bio (
                                     username character varying(24) NOT NULL,
                                     bio text NOT NULL
);


ALTER TABLE "issue#1".users_bio OWNER TO "issue#1_dev";

--
-- Name: feeds id; Type: DEFAULT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feeds ALTER COLUMN id SET DEFAULT nextval('"issue#1".feeds_id_seq'::regclass);


--
-- Name: posts id; Type: DEFAULT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts ALTER COLUMN id SET DEFAULT nextval('"issue#1".post_id_seq'::regclass);


--
-- Name: releases id; Type: DEFAULT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases ALTER COLUMN id SET DEFAULT nextval('"issue#1".title_id_seq'::regclass);


--
-- Data for Name: channel_admins; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".channel_admins VALUES ('chromagnum', 'rembrandt', true);
INSERT INTO "issue#1".channel_admins VALUES ('chromagnum', 'loveless', false);
INSERT INTO "issue#1".channel_admins VALUES ('IsisCane', 'IsisCane', true);
INSERT INTO "issue#1".channel_admins VALUES ('haroma', 'haroma', true);
INSERT INTO "issue#1".channel_admins VALUES ('faberge', 'faberge', true);


--
-- Data for Name: channel_official_catalog; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".channel_official_catalog VALUES ('chromagnum', 6, 4);
INSERT INTO "issue#1".channel_official_catalog VALUES ('chromagnum', 53, 5);
INSERT INTO "issue#1".channel_official_catalog VALUES ('chromagnum', 73, 6);


--
-- Data for Name: channel_pictures; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- Data for Name: channel_stickies; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- Data for Name: channels; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".channels VALUES ('chromagnum', '2019-12-29 19:59:22.564274+03', 'THE FUTURE IS NOW', 'take it off');
INSERT INTO "issue#1".channels VALUES ('tempOne', '2020-01-22 20:19:15.71383+03', 'tempOne''s channel', NULL);
INSERT INTO "issue#1".channels VALUES ('IsisCane', '2020-01-16 21:44:42.22749+03', 'Isis Cane''s channel', NULL);
INSERT INTO "issue#1".channels VALUES ('haroma', '2020-01-26 19:42:52.97117+03', 'haroma''s channel', NULL);
INSERT INTO "issue#1".channels VALUES ('faberge', '2020-01-26 20:10:16.561102+03', 'faberge''s channel', NULL);


--
-- Data for Name: comments; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 3, -1, '<p>they can hear me crying… </p>
', 'loveless', '2020-01-11 20:16:33.519527+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 8, -1, 'It''s easy if you''re passionate...', 'loveless', '2020-01-15 02:14:45.403771+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 1, -1, 'Epistien did not kill himself!', 'slimmy', '2020-01-11 19:29:29.058057+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 2, 1, 'duhh', 'rembrandtian', '2020-01-11 19:30:04.273774+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 6, 3, '<p>the diamond mines are <em>burning</em> </p>
', 'rembrandtian', '2020-01-11 20:26:34.222721+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 17, -1, 'onlyfans.com/jQueryKink', 'rembrandt', '2020-01-28 23:00:07.207603+03');


--
-- Data for Name: feed_subscriptions; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".feed_subscriptions VALUES (3, 'chromagnum', '2020-01-06 17:01:08.862048+03');
INSERT INTO "issue#1".feed_subscriptions VALUES (6, 'chromagnum', '2020-01-23 00:00:24.744977+03');
INSERT INTO "issue#1".feed_subscriptions VALUES (7, 'chromagnum', '2020-01-23 00:00:30.243607+03');
INSERT INTO "issue#1".feed_subscriptions VALUES (8, 'chromagnum', '2020-01-23 00:00:35.248254+03');
INSERT INTO "issue#1".feed_subscriptions VALUES (3, 'IsisCane', '2020-01-23 00:34:50.41627+03');


--
-- Data for Name: feeds; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".feeds VALUES ('rembrandt', 'hot', 3);
INSERT INTO "issue#1".feeds VALUES ('slimmy', 'top', 7);
INSERT INTO "issue#1".feeds VALUES ('loveless', 'new', 5);
INSERT INTO "issue#1".feeds VALUES ('rembrandtian', 'top', 6);
INSERT INTO "issue#1".feeds VALUES ('IsisCane', 'hot', 8);
INSERT INTO "issue#1".feeds VALUES ('haroma', 'hot', 15);
INSERT INTO "issue#1".feeds VALUES ('faberge', 'hot', 16);


--
-- Data for Name: post_contents; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".post_contents VALUES (5, 53);
INSERT INTO "issue#1".post_contents VALUES (6, 73);


--
-- Data for Name: post_stars; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- Data for Name: posts; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".posts VALUES (4, 'tipstitipst[tipstst]tipstitstipstsi[ps]', 'HEREWEGOLANG', 'rembrandt', 'chromagnum', '2020-01-04 20:58:27.926859+03');
INSERT INTO "issue#1".posts VALUES (5, NULL, 'cordovadovadovadovedova', 'rembrandt', 'chromagnum', '2019-12-29 20:00:53.714587+03');
INSERT INTO "issue#1".posts VALUES (3, 'B7 Chord, G6 Chord.', 'its so strangeeee', 'loveless', 'chromagnum', '2019-12-29 19:59:39.568527+03');
INSERT INTO "issue#1".posts VALUES (6, 'Welcome!', 'Issue #1 v0.1', 'rembrandt', 'chromagnum', '2020-01-12 18:54:48.460537+03');


--
-- Data for Name: release_metadata; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".release_metadata VALUES (73, 'Here it is folks.', '{"genres": ["Programming", "Design"], "authors": ["Yohe-Am"]}', 'Coding', '2020-01-28 13:14:16.017584+03', 'ISSUE #1 DESIGN DOC');
INSERT INTO "issue#1".release_metadata VALUES (6, 'tips baby, tips', '{"genres": ["School"], "authors": ["@rembrandt"]}', 'Guide', '2020-01-04 11:25:15.055178+03', 'Tips1.md');
INSERT INTO "issue#1".release_metadata VALUES (53, 'a broken moon', '{"genres": ["Cartoon"], "authors": ["Rebbecca Sugar"]}', 'Cartoon', '2016-01-04 23:06:16.017584+03', 'Please don''t leave Pink...don''t leave please.');
INSERT INTO "issue#1".release_metadata VALUES (52, 'Forget-this-not.', '{"genres": ["Omnious Message", "Prophecy"], "authors": ["Hooded Messenger"]}', 'Message', '2020-01-05 13:14:16.017584+03', 'The Journey Ends!');
INSERT INTO "issue#1".release_metadata VALUES (54, 'Guidelines to the future and to the deep, honest, archaic path.', '{"genres": ["K-Pop"], "authors": ["Hooded Messenger"]}', 'Literature', '2020-12-05 13:14:16.017584+03', 'Above & Not There Yet.');
INSERT INTO "issue#1".release_metadata VALUES (68, 'Full stop.', '{"genres": ["Catastrophe"], "authors": ["Man"]}', 'Atomic', '2020-01-24 23:16:09.273085+03', 'This is Not A Test');


--
-- Data for Name: releases; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".releases VALUES (6, 'chromagnum', 'text', '2020-01-04 11:23:04.271553+03');
INSERT INTO "issue#1".releases VALUES (52, 'chromagnum', 'text', '2020-01-05 13:18:43.7471+03');
INSERT INTO "issue#1".releases VALUES (53, 'chromagnum', 'image', '2020-01-05 13:55:05.730832+03');
INSERT INTO "issue#1".releases VALUES (54, 'chromagnum', 'text', '2020-01-12 18:05:22.383129+03');
INSERT INTO "issue#1".releases VALUES (68, 'chromagnum', 'text', '2020-01-24 23:16:09.274719+03');
INSERT INTO "issue#1".releases VALUES (73, 'chromagnum', 'text', '2020-01-28 23:07:26.363033+03');


--
-- Data for Name: releases_image_based; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".releases_image_based VALUES (53, 'release.1eb2969c-cf4c-40cb-9b78-39bea2b089da.the image.jpg');


--
-- Data for Name: releases_text_based; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".releases_text_based VALUES (52, 'THE EDGE BECKONS! Heed it''s call at 27 o''clock.');
INSERT INTO "issue#1".releases_text_based VALUES (54, 'The wind ruffles my feathers. Yet I do not see no blue sky...');
INSERT INTO "issue#1".releases_text_based VALUES (6, '
# Guideline

## Project setup

### git cloning
- open Git Bash
```
cd ~/go/src/github.com
mkdir slim-crown
git clone https://github.com/slim-crown/issue-1-REST.git
git clone https://github.com/slim-crown/issue-1-website.git
```
	- if you are unlucky enough not to have Wi-Fi at home, ask me. I have
	tools that help.

- how to create your own branch from the next branch
```
cd ~/go/src/github.com/slim-crown/issue-1-REST
git checkout -b feature-user-service origin/next  ## the new branch is called `feature-user-service`
```
- how to merge work done on a branch found on github to your branch
```
cd ~/go/src/github.com/slim-crown/issue-1-REST
git checkout feature-user-service			## open your branch first
git merge origin/feature-feed-service    	## you merged work done from the feed-sercvice branch into your own
```


### dependencies
`go get https://github.com/gorilla/mux`

### sql - obsolete
open pgAdmin
	- you need to Run As Administrator
	- it usually fails to open on the first attempt so try until it does
open Query Tool
	- can be found on the menu bar under Tools
copy text from *globals.sql* to query text area
run query
	- lighting symbol on the tool bar
do same for *issue#1.sql*
	- order is essential

## Project tips
**READ THE WHOLE THING, DON''T STOP MIDWAY.**
First of all, read the *Succesful Git Branching* article I sent to create a new branch and do work on. Don''t go ahead without reading that article.
Ask if you have any questions.

And also, use a simple text file to design and plan things ahead.

After you thing you have a basic understanding of the project structure and you feel confident enough to start work, start from the handlers *endpoints* for the service you''re going to implement.
Just list out the URL''s you expect to handled
e.g
- POST		/users // create user from given JSON or Form
- GET 		/users/{username} // get user at username in JSON
- PUT	 	/users/{username} // update/create user at username from given JSON
- DELETE	/users/{username}
and *especially* GET on all users which also includes our search option
- GET 		/users?sortBy=creation-time&sortOrder=asc&limit=25&offset=0&pattern=John

Implement the handlers first and as you go one, you''ll soon have a list of methods that you need your service to have which should make things easy. Go on and *specify* the methods on your Service *inteface*.

Handlers are one of the toughest part to implement, with tricky logic and lots of error handling so be sure to look at the functions I did and ask questions about any constructs you don''y understand.

When you''re done with the handlers, it would be a good time to commit. Since its *a really sucky thing to do* commiting a project that doesn''t compile, make sure to define all the Service methods that you used on the handlers for your Service interface and fix any other errors before you do. I really reccommend getting done with the handlers now, I beleive its much easier that way. Just make sure the project runs before you commit.
Ask if you have any questions.

Now it''s time to do the Service. By now, you should have a list of methods that you should implement and your Service interface should may look something like this:
```
// Service specifies a method to service User entities.
type Service interface {
	AddUser(user *User) error
	GetUser(username string) (*User, error)
	UpdateUser(username string, u *User) error
	DeleteUser(username string) error
	SearchUser(pattern, sortBy, sortOrder string, limit, offset int) ([]*User, error)
	PassHashIsCorrect(username, passHash string) (bool, error)
	BookmarkPost(username string, postID int) error
}
```

Did you notice ther `PassHashIsCorrect` & `BookmarkPost` methods?
Besides the basic CRUD methods, your entites will most likely require special methods that they need to function. Try to think of some and discuss with mates.

Now, like you did with the Handlers, implement all the methods on the Service without worrying about the Repository and *defining* methods that you need on the Repository interface along the way.

Services should be the easiest to implement as most methods are simple calls on the Repository that don''t do much besides input validation and stuff.

Like before, now is also a good time to commit. Make sure the project runs and you''ve specifed all the Repository methods that you used.

Next on, go to the memory implementation of the Repository. You''ll be implementing the Repository inteface from earler using simple in memory/cache. This is also very easy as most methods are direct calls to the DB based implementation of the same Repository interface with a little caching in between.

Make sure the project runs, commit.

Implement the same Repository using PostgreSQL.
This will most likely be the hardest step I''m afraid where you might end up with a lot of errors. But, since lucky you did the handlers and the sercvice and stuff first, you''ll be able to test your methods right there and then.

Do that, **Create a pull request.**, boom. You''re done.


Note: ofcourse you''re not completely done. There will be different specialized methods that each service will have to handle (for e.g. like authenticate for the user service) that you''ll have to work on but that''s for later.

Use the following setup to plan your work.

- [x] Service methods
	- [x] Get User
	- [x] Get Users
	- [x] Update User
	- [x] Delete User
	- [x] Search User
	- [x] Authenticate User
- [x] Cache Repository methods
	- [x] Get User
	- [x] Get Users
	- [x] Update User
	- [x] Delete User
	- [x] Search User
	- [x] Authorize User
- [ ] Postgre Repository methods
	- [x] Get User
	- [x] Get Users
	- [x] Update User
	- [x] Delete User
	- [ ] Search User
	- [x] Authorize User
- [ ] HTTP handler functions
	- [x] GET /users/{username}
	- [x] GET /users
	- [x] POST /users
	- [x] PUT /users/{username}
	- [x] DELETE /users{username}
	- [X] POST /users/{username}?authenticate

## Other stuff
Get done with this shit ***fast*** as the REST API is only half the work and according to the syllabus....
-------------------------------------------------------------
Week 11 | 	Securing Web Applications 	 |	Add security features
-------------------------------------------------------------
Week 12 | 	Testing Web Applications  	 |	Test web applications
-------------------------------------------------------------
Week 13 |	Concurrency for Performance  |	Improve web application
		|	Improvement					 |	performance by using
		|								 |	concurrency
-------------------------------------------------------------
Week 14 |	Internationalization and	 |	Use internationalization
		|	Localization				 |	and localization

... who know how much of these he might ask us to implememt.

Also, project evaluation I is 20% compared to Project Evaluation II which is 10%.
With such a big project as ours, we''re hard pressed to deliver our features by the first deadline.');
INSERT INTO "issue#1".releases_text_based VALUES (68, '1 minute and 39 seconds to midnight.');
INSERT INTO "issue#1".releases_text_based VALUES (73, '# Entities
- Publication
  - an atomic work of creativity supported by HTML 1.1 (hopefully)
- User
- Channel
  - A place where creators release to the world
- Post
  - Contains a publication(s), has like buttons, has comment board. Gets born in achannel.
- Comment
- Feed
  - Channels Channels'' posts into a single list of posts sorted based on...something.

# Actions
- Users
	- add
	- modify
	- remove
- Release
	- add
	- edit
	- modify Metadata
	- view
	- remove
- Channels
	- add
	- modify channel
	- modify admins
	- view Channel
	- view catalog
	- remove
	- subscribe
- Post
	- add Post
		- attach release
		- catalog attached release
	- star post
	- sticky post
	- bookmark post
	- view Post
	- view attached release
	- view Comments
	- remove
- Comment
	- add
	- view Comment
	- view Replies
	- remove
- Feed
	- view
	- change sorting
- Search
	- search by string

# Expected Webshite URLs
```
~/c/channelname
~/p/postID
~/r/releaseid
~/u/username
```

*** dead men don''t cry ***
__ waah gwann __
* fellla ya got it good *
_fellla got it good_
dead **men don''t** cry


# TODO
- [x] image server
- [x] auth
- [x] search
- [-] sanitizers
- [ ] session
- [ ] website
- [ ] handler tests
- [ ] repo mocks

- [x] validate email
- [x] migrate to stdlib logger
- [x] migrage to bcrypt
- [x] auto create channel and feed during user creation
- [x] make user owner of auto created channel
- [ ] sorting for comments
- [x] test for invalid characters inside usernames
# Bugs
- [x] check for password length
- [ ] pre-sort all by `creation_time` after search
- [ ] diffrenciate between put and patch
- [x] addUser postgre set email/password to zero
- [x] sort by last_name first_name bug
- [ ] delete all solely adminstrated channels when user removed
- [x] fix refresh tokens function
- [ ] limit the limit

- [ ] make feed service return data on POST, PUT
- [ ] return 204 for no update POST
- [ ] return 204 for DELETE
- [ ] return no of subscribed channels along with feed
- [x] username length is 24 chars

## web app

- [ ] Input Validation
- [ ] Pages
  - [x] Front
  - [x] Signup Form
  - [ ] Login Form
  - [ ] Navbar !
    - [ ] List Subscription Button
    - [ ] Search Navbar Text Input
    - [ ] Change Feed Sorting Navbar Button
  - [ ] Home
    - [ ] Feed PostList !
  - [ ] Add Channel Form
  - [ ] Add/Edit Post Form !
  - [ ] Post View
    - [ ] ReleseView Component  !
    - [ ] Comment Board (Add/Edit Comment) !
  - [ ] Channel View
	- [ ] PostList Tab !
	- [ ] Official Releases Tab !
	- [ ] Releses Tab
	- [ ] Edit Channel Form
	- [ ] Add/Edit Release Form !
    - [ ] Profile Tab (Personal Channels) !
  - [ ] SearchView !

## session implementation
- [ ] session
  - [x] Expire sessions
  - [ ] Check if user has a previous session
  - [x] new sesison on login
  - [ ] Cache repo
  - [x] Gorm Repo

## auth implementation
- [x] HTTP handler functions
  - [x] POST /token-auth
  - [x] GET  /token-auth-refresh
  - [x] GET  /logout
- [x] JWT backend methods

## Search service implementation

- [x] HTTP handler functions
  - [x] GET  /search
- [x] Service methods
  - [x] Search Comments
- [x] Postgre Repository methods
	- [x] Search Comments

## User Service implementation
- [x] HTTP handler functions
	- [x] GET /users/{username}
	- [x] GET /users
	- [x] POST /users
	- [x] PUT /users/{username}
	- [x] DELETE /users{username}
	- [x] GET /users/{username}/bookmarks
	- [x] PUT /users/{username}/bookmarks
	- [x] DELETE /users/{username}/bookmarks/{postID}
	- [x] POST /users/{username}/bookmarks
	- [x] GET /users/{username}/picture
	- [x] DELETE /users/{username}/picture
	- [x] PUT /users/{username}/picture
- [x] Service methods
	- [x] Get User
	- [x] Get Users
	- [x] Update User
	- [x] Delete User
	- [x] Search User
	- [x] Authenticate User
	- [x] Bookmark Post
	- [x] Delete Bookmark
	- [x] Add Picture
	- [x] Remove Picture
- [x] Cache Repository methods
	- [x] Get User
	- [x] Get Users
	- [x] Update User
	- [x] Delete User
	- [x] Search User
	- [x] PassHashIsCorrect
	- [x] Bookmark Post
	- [x] Delete Bookmark
- [x] Postgre Repository methods
	- [x] Get User
	- [x] Get Users
	- [x] Update User
	- [x] Delete User
	- [x] Search User
	- [x] PassHashIsCorrect User
	- [x] Bookmark Post
	- [x] Delete Bookmark
	- [x] Add Picture
	- [x] Remove Picture

## Feed Service implementation
- [x] HTTP handler functions
	- [x] GET /users/{username}/feed/
	- [x] GET /users/{username}/feed/posts?sort=new
	- [x] GET /users/{username}/feed/channels
	- [x] POST /users/{username}/feed/channels
	- [x] PUT /users/{username}/feed
	- [x] DELETE /users/{username}/feed/channels/{username}
- [x] Service methods
	- [x] Get Feed
	- [x] Get Feed Channels
	- [x] Get Feed Posts
	- [x] Update Feed
	- [x] Delete Feed
	- [x] Subscribe to Channel
	- [x] Unbscribe from Channel
- [x] Cache Repository methods
	- [x] Get Feed
	- [x] Get Feed Channels
	- [x] Get Feed Posts
	- [x] Update Feed
	- [x] Delete Feed
	- [x] Subscribe to Channel
	- [x] Unbscribe from Channel
- [x] Postgre Repository methods
	- [x] Get Feed
	- [x] Get Feed Channels
	- [x] Get Feed Posts
	- [x] Update Feed
	- [x] Delete Feed
	- [x] Subscribe to Channel
	- [x] Unbscribe from Channel

## Release Service implementation
- [x] HTTP handler functions
	- [x] GET `/releases/{id}`
	- [x] GET `/releases?sort=time`
	- [x] POST `/releases`
	- [x] PUT `/releases/{id}`
	- [x] DELETE `/releases/{id}`
- [x] Service methods
	- [x] Get Release
	- [x] Add Release
	- [x] Update Release
	- [x] Delete Release
- [x] Cache Repository methods
	- [x] Get Release
	- [x] Add Release
	- [x] Update Release
	- [x] Delete Release
- [x] Postgre Repository methods
	- [x] Get Release
	- [x] Add Release
	- [x] Update Release
	- [x] Delete Release


## Comment Service implementation
- [x] HTTP handler functions
	- [x] GET `/posts/{postID}/comments/{commentID}`
	- [x] GET `/posts/{postID}/comments?sort=time`
	- [x] POST `/posts/{postID}/comments`
	- [x] PATCH `/posts/{postID}/comments/{commentID}`
	- [x] DELETE `/posts/{postID}/comments/{commentID}`
	- [x] GET `/posts/{postID}/comments/{rootCommentID}/replies/{commentID}`
	- [x] GET `/posts/{postID}/comments/{rootCommentID}/replies?sort=time`
	- [x] POST `/posts/{postID}/comments/{rootCommentID}/replies`
	- [x] PATCH `/posts/{postID}/comments/{rootCommentID}/replies/{commentID}`
	- [x] DELETE `/posts/{postID}/comments/{rootCommentID}/replies/{commentID}`
- [x] Service methods
	- [x] Get Comment
	- [x] Add Comment
	- [x] Update Comment
	- [x] Delete Comment
- [x] Cache Repository methods
	- [x] Get Comment
	- [x] Add Comment
	- [x] Update Comment
	- [x] Delete Comment
- [x] Postgre Repository methods
	- [x] Get Comment
	- [x] Add Comment
	- [x] Update Comment
	- [x] Delete Comment


```shell

"C:\Program Files\PostgreSQL\12\bin\pg_dump.exe" "--file=C:\Users\cosmicbridgeman\Desktop\dev\issue-1\db\db dump\issue#1_db.sql" --create --inserts --dbname=issue#1_db --username=issue#1_dev --host=localhost --port=5432
```

```sql
SELECT username, COALESCE(first_name, ''''), COALESCE(middle_name, ''''), COALESCE(last_name, ''''), creation_time
FROM users
WHERE username ~~* ''%New%'' OR first_name ~~* ''%New%'' OR last_name ~~* ''%New%''
-- SELECT * from users
-- where to_tsvector(''simple'', concat_ws('' '',username,first_name,last_name,middle_name)) @@ to_tsquery(''rem'')
```


- [x] Database Tables (Schemas)
- [x] Repository (for CRUD operation)
- [x] Service (for CRUD operation)
- [x] Handlers (that handle CRUD requests)
- [ ] Template pages (for each CRUD operation)
Testing
- [ ] Mocking at least the repository
- [ ] Testing at least the handlers
Security
    - [ ] Input Validation of all inputs
    - [x] Authentication
    - [ ] Authorization 
    - [ ] CSRF Protection
    - [x] SQL Injection Protection
    - [x] Session Cookie Management
REST API
    - [x] Routes follow RESTful API design guidelines
    - [x] HTTP responses show proper messages and status code
Deployment
    - [x] Local/Standalone Deployment
    - [ ] Docker Deployment (Local and on DockerHub)
    - [ ] Cloud Deployment (Choose any platform from available cloud providers)
- [x] Code Structure Your project follows the Clean Architecture
- [x] Git History At least 30 commit (each member at least 10)
- [ ] Using Kanban Progress tracking in Kanban

Demo
	- [ ] Feature one Showing CRUD Operation
	- [ ] Feature two Showing CRUD Operation 
	- [ ] Feature three Showing CRUD Operation
	- [ ] Feature Four Showing CRUD Operation
	- [ ] Security
	- [ ] Show Input Validation
	- [ ] Access protected resource without login
	- [ ] Delete Client Session and access protected resource
	- [ ] Login then access protected resources and logout then access protected resources
	- [ ] Show role based permission   
	- [ ] Show CSRF prevention   
	- [ ] Testing Run test and show the result   
	REST API 
		- [ ] Do CRUD operation for one of your features    ');


--
-- Data for Name: tsvs_comment; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".tsvs_comment VALUES (3, '''cri'':5 ''hear'':3 ''loveless'':6B');
INSERT INTO "issue#1".tsvs_comment VALUES (8, '''easi'':3 ''loveless'':8B ''passion'':7 ''re'':6');
INSERT INTO "issue#1".tsvs_comment VALUES (1, '''epistien'':1 ''kill'':4 ''slimmi'':5B');
INSERT INTO "issue#1".tsvs_comment VALUES (2, '''duhh'':1 ''rembrandtian'':2B');
INSERT INTO "issue#1".tsvs_comment VALUES (6, '''burn'':5 ''diamond'':2 ''mine'':3 ''rembrandtian'':6B');
INSERT INTO "issue#1".tsvs_comment VALUES (17, '''/jquerykink'':3 ''onlyfans.com'':2 ''onlyfans.com/jquerykink'':1 ''rembrandt'':4B');


--
-- Data for Name: tsvs_posts; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".tsvs_posts VALUES (4, '''chromagnum'':7B ''herewegolang'':5A ''ps'':4 ''rembrandt'':6B ''tipstitipst'':1 ''tipstitstipstsi'':3 ''tipstst'':2');
INSERT INTO "issue#1".tsvs_posts VALUES (5, '''chromagnum'':3B ''cordovadovadovadovedova'':1A ''rembrandt'':2B');
INSERT INTO "issue#1".tsvs_posts VALUES (3, '''b7'':1 ''chord'':2,4 ''chromagnum'':9B ''g6'':3 ''loveless'':8B ''strangeee'':7A');
INSERT INTO "issue#1".tsvs_posts VALUES (6, '''1'':3A ''chromagnum'':6B ''issu'':2A ''rembrandt'':5B ''v0.1'':4A ''welcom'':1');


--
-- Data for Name: tsvs_release; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".tsvs_release VALUES (73, '''-1'':917 ''/bookmarks'':452,457,462,468 ''/c/channelname'':134 ''/comments'':789,795,802,807,813,819,827,836,843,851 ''/feed'':589,611 ''/feed/channels'':601,606,616 ''/feed/posts'':594 ''/logout'':395 ''/p/postid'':135 ''/picture'':473,478,483 ''/posts'':787,793,800,805,811,817,825,834,841,849 ''/r/releaseid'':136 ''/releases'':713,717,722,725,729 ''/replies'':821,829,838,845,853 ''/search'':409 ''/token-auth'':389 ''/token-auth-refresh'':392 ''/u/username'':137 ''/users'':432,436,439,442,446,450,455,460,466,471,476,481,587,592,599,604,609,614 ''1'':15A,927,931 ''1.1'':28 ''10'':1096 ''12'':907 ''1_db.sql'':922 ''204'':273,279 ''24'':294 ''30'':1090 ''5432'':936 ''access'':1128,1137,1142,1148 ''achannel'':54 ''action'':70 ''add'':72,76,83,95,117,324,512,573,738,754,770,862,878,894 ''add/edit'':327,336,350 ''addus'':234 ''admin'':87 ''adminstr'':251 ''along'':287 ''api'':1040,1045,1166 ''app'':297 ''architectur'':1084 ''atom'':21 ''attach'':97,100,111 ''auth'':164,381 ''authent'':503,1027 ''author'':1028 ''auto'':187,200 ''avail'':1073 ''backend'':398 ''base'':67,1153 ''bcrypt'':185 ''bin'':908 ''board'':50,335 ''bookmark'':106,506,510,539,543,567,571 ''born'':52 ''bug'':213,247 ''button'':47,311,320 ''c'':903,911 ''cach'':376,518,647,747,871 ''catalog'':91,99 ''chang'':125,316 ''channel'':31,57,58,82,85,89,189,202,252,286,325,338,348,356,627,641,645,656,670,674,685,699,703 ''char'':295 ''charact'':210 ''check'':215,364 ''choos'':1069 ''chromagnum'':1B ''clean'':1083 ''client'':1134 ''cloud'':1067,1074 ''coalesc'':940,943,946 ''code'':13B,1056,1077 ''comment'':49,55,114,116,119,205,334,337,415,422,778,860,863,866,869,876,879,882,885,892,895,898,901 ''commentid'':790,808,814,822,846,854 ''commit'':1091 ''compon'':333 ''concat'':971 ''contain'':41 ''cooki'':1037 ''cosmicbridgeman'':913 ''creat'':188,201,923 ''creation'':194,224,949 ''creativ'':24 ''creator'':35 ''cri'':142,158 ''crud'':990,995,1001,1007,1107,1112,1117,1122,1168 ''csrf'':1029,1156 ''data'':268 ''databas'':984 ''db'':918,919,928 ''dbname'':925 ''dead'':138,154 ''delet'':248,281,445,459,475,497,509,531,542,558,570,613,636,665,694,728,744,760,776,810,848,868,884,900,1133 ''demo'':1103 ''deploy'':1057,1060,1062,1068 ''design'':9C,16A,1046 ''desktop'':914 ''dev'':915,932 ''diffrenci'':228 ''doc'':17A ''docker'':1061 ''dockerhub'':1066 ''dump'':920 ''edit'':77,347 ''email'':176 ''email/password'':237 ''entiti'':18 ''expect'':131 ''expir'':362 ''featur'':1104,1109,1114,1119,1174 ''feed'':56,123,191,265,289,317,322,578,623,626,630,634,637,652,655,659,663,666,681,684,688,692,695 ''fellla'':145,150 ''file'':905,910 ''first'':245,941,957,974 ''fix'':257 ''folk'':6 ''follow'':1043,1081 ''form'':305,307,326,329,349,352 ''four'':1120 ''front'':302 ''function'':260,386,406,429,584,710,784 ''get'':51,391,394,408,431,435,449,470,488,491,522,525,549,552,586,591,598,622,625,629,651,654,658,680,683,687,712,716,735,751,767,786,792,816,824,859,875,891 ''git'':1086 ''good'':149,153 ''gorm'':379 ''got'':147,151 ''guidelin'':1047 ''gwann'':144 ''handl'':1000 ''handler'':170,385,405,428,583,709,783,998,1019 ''histori'':1087 ''home'':321 ''hope'':29 ''host'':933 ''html'':27 ''http'':384,404,427,582,708,782,1049 ''id'':714,726,730 ''imag'':161 ''implement'':359,382,402,425,580,706,780 ''inject'':1033 ''input'':298,315,1021,1025,1126 ''insert'':924 ''insid'':211 ''invalid'':209 ''issu'':14A,916,921,926,930 ''jwt'':397 ''kanban'':1098,1102 ''last'':243,947,961,976 ''least'':1012,1017,1089,1095 ''length'':218,292 ''like'':46 ''limit'':261,263 ''list'':63,309 ''local'':1063 ''local/standalone'':1059 ''localhost'':934 ''logger'':181 ''login'':306,375,1132,1140 ''logout'':1146 ''make'':196,264 ''manag'':1038 ''member'':1093 ''men'':139,155 ''messag'':1053 ''metadata'':79 ''method'':399,412,419,486,520,547,620,649,678,733,749,765,857,873,889 ''middl'':944,978 ''migrag'':183 ''migrat'':178 ''mock'':173,1010 ''modifi'':73,78,84,86 ''name'':244,246,942,945,948,958,962,975,977,979 ''navbar'':308,313,319 ''new'':372,596,955,959,963 ''offici'':342 ''one'':1105,1171 ''oper'':991,996,1008,1108,1113,1118,1123,1169 ''owner'':198 ''page'':300,1004 ''passhashiscorrect'':537,564 ''password'':217 ''patch'':232,804,840 ''permiss'':1154 ''person'':355 ''pg_dump.exe'':909 ''pictur'':513,516,574,577 ''place'':33 ''platform'':1071 ''port'':935 ''post'':40,59,65,94,96,103,105,107,109,270,277,328,330,388,438,465,507,540,568,603,631,660,689,721,799,833 ''postgr'':235,417,545,676,763,887 ''postgresql'':906 ''postid'':463,788,794,801,806,812,818,826,835,842,850 ''postlist'':323,340 ''pre'':220 ''pre-sort'':219 ''prevent'':1157 ''previous'':369 ''profil'':353 ''program'':7C,904 ''progress'':1099 ''project'':1080 ''proper'':1052 ''protect'':1030,1034,1129,1138,1143,1149 ''provid'':1075 ''public'':19,43 ''put'':230,271,441,454,480,608,724 ''refresh'':258 ''releas'':36,75,98,101,112,343,351,704,736,739,742,745,752,755,758,761,768,771,774,777 ''reles'':345 ''releseview'':332 ''rem'':982 ''remov'':74,81,92,115,122,255,515,576 ''repli'':121 ''repo'':172,377,380 ''repositori'':418,519,546,648,677,748,764,872,888,988,1014 ''request'':1002 ''resourc'':1130,1139,1144,1150 ''respons'':1050 ''rest'':1039,1044,1165 ''result'':1164 ''return'':267,272,278,282 ''role'':1152 ''rootcommentid'':820,828,837,844,852 ''rout'':1042 ''run'':1159 ''sanit'':167 ''schema'':986 ''search'':127,128,166,227,312,400,414,421,500,534,561 ''searchview'':357 ''secur'':1020,1124 ''select'':938,964 ''server'':162 ''servic'':266,401,411,424,485,579,619,705,732,779,856,993 ''sesison'':373 ''session'':168,358,360,363,370,1036,1135 ''set'':236 ''shell'':902 ''show'':1051,1106,1111,1116,1121,1125,1151,1155,1162 ''signup'':304 ''simpl'':970 ''singl'':62 ''sole'':250 ''someth'':69 ''sort'':66,126,203,221,241,318,595,718,796,830 ''sql'':937,1032 ''star'':102 ''status'':1055 ''stdlib'':180 ''sticki'':104 ''string'':130 ''structur'':1078 ''subscrib'':93,285,639,668,697 ''subscript'':310 ''support'':25 ''tab'':341,344,346,354 ''tabl'':985 ''templat'':1003 ''test'':171,207,1009,1015,1158,1160 ''text'':2A,314 ''three'':1115 ''time'':225,719,797,831,950 ''todo'':159 ''token'':259 ''track'':1100 ''tsqueri'':981 ''tsvector'':969 ''two'':1110 ''unbscrib'':643,672,701 ''updat'':276,494,528,555,633,662,691,741,757,773,865,881,897 ''url'':133 ''use'':1097 ''user'':30,71,193,197,254,366,423,489,492,495,498,501,504,523,526,529,532,535,550,553,556,559,562,565,912,952,966 ''usernam'':212,291,433,443,447,451,456,461,467,472,477,482,588,593,600,605,610,615,617,929,939,954,973 ''valid'':175,299,1022,1127 ''view'':80,88,90,108,110,113,118,120,124,331,339 ''waah'':143 ''web'':296 ''webshit'':132 ''websit'':169 ''without'':1131 ''work'':22 ''world'':39 ''ws'':972 ''x'':160,163,165,174,177,182,186,195,206,214,233,240,256,290,301,303,361,371,378,383,387,390,393,396,403,407,410,413,416,420,426,430,434,437,440,444,448,453,458,464,469,474,479,484,487,490,493,496,499,502,505,508,511,514,517,521,524,527,530,533,536,538,541,544,548,551,554,557,560,563,566,569,572,575,581,585,590,597,602,607,612,618,621,624,628,632,635,638,642,646,650,653,657,661,664,667,671,675,679,682,686,690,693,696,700,707,711,715,720,723,727,731,734,737,740,743,746,750,753,756,759,762,766,769,772,775,781,785,791,798,803,809,815,823,832,839,847,855,858,861,864,867,870,874,877,880,883,886,890,893,896,899,983,987,992,997,1026,1031,1035,1041,1048,1058,1076,1085 ''ya'':146 ''yohe'':12C ''yohe-am'':11C ''zero'':239');
INSERT INTO "issue#1".tsvs_release VALUES (6, '''/go/src/github.com'':20 ''/go/src/github.com/slim-crown/issue-1-rest'':66,99 ''/gorilla/mux'':131 ''/slim-crown/issue-1-rest.git'':29 ''/slim-crown/issue-1-website.git'':34 ''/users'':294,303,312,322,337,948,952,955,958,962,966 ''0'':347 ''1.sql'':188 ''10'':1054 ''11'':992 ''12'':1000 ''13'':1008 ''14'':1021 ''20'':1046 ''25'':345 ''abl'':813 ''accord'':987 ''add'':996 ''addus'':555 ''administr'':141 ''afraid'':786 ''ahead'':225,248 ''along'':659 ''also'':237,331,692,739,1041 ''api'':980 ''applic'':995,1003,1006,1014 ''area'':175 ''articl'':210,229 ''asc'':343 ''ask'':48,230,417,507,1037 ''attempt'':150 ''authent'':855,899,968 ''author'':921,941 ''b'':69 ''babi'':4 ''bar'':165,183 ''base'':751 ''bash'':18 ''basic'':255,605 ''beleiv'':492 ''besid'':603,683 ''big'':1058 ''bookmarkpost'':590,601 ''bool'':588 ''boom'':828 ''branch'':60,64,77,91,97,108,122,209,217 ''cach'':761,902 ''call'':79,674,747 ''cd'':19,65,98 ''checkout'':68,101 ''chromagnum'':1B ''clone'':15,26,31 ''commit'':440,449,506,697,769 ''compar'':1047 ''compil'':455 ''complet'':837 ''concurr'':1009,1019 ''confid'':264 ''construct'':421 ''copi'':168 ''creat'':57,214,295,824 ''creation'':340 ''creation-tim'':339 ''crown'':24 ''crud'':606 ''db'':750 ''deadlin'':1073 ''defin'':459,650 ''delet'':321,893,915,936,961 ''deleteus'':570 ''deliv'':1067 ''depend'':126 ''design'':244 ''differ'':842 ''direct'':746 ''discuss'':627 ''doesn'':453 ''done'':88,116,429,486,831,838,972 ''e.g'':292,853 ''earler'':732 ''easi'':378,741 ''easier'':495 ''easiest'':666 ''end'':790 ''endpoint'':273 ''enough'':39,265 ''entit'':609 ''entiti'':551 ''error'':404,478,558,563,569,573,583,589,595,796 ''especi'':325 ''essenti'':191 ''evalu'':1043,1050 ''expect'':289 ''fail'':144 ''fast'':976 ''featur'':71,81,103,998,1069 ''feature-user-servic'':70,80,102 ''feed'':120 ''feed-sercvic'':119 ''feel'':263 ''fi'':45 ''file'':242 ''first'':109,149,202,353,809,1072 ''fix'':475 ''follow'':874 ''form'':301 ''found'':92,161 ''function'':413,620,945 ''get'':128,302,305,326,336,485,884,887,906,909,927,930,947,951,971 ''getus'':559 ''git'':14,17,25,30,67,100,110,208 ''github'':94 ''github.com'':28,33,130 ''github.com/gorilla/mux'':129 ''github.com/slim-crown/issue-1-rest.git'':27 ''github.com/slim-crown/issue-1-website.git'':32 ''given'':298,319 ''globals.sql'':171 ''go'':127,224,279,357,379,716 ''good'':437,694 ''guid'':9B ''guidelin'':11 ''half'':983 ''handl'':291,405,851 ''handler'':272,352,389,432,469,489,636,803,944 ''hard'':1064 ''hardest'':782 ''help'':54 ''home'':47 ''http'':943 ''ii'':1051 ''implememt'':1040 ''implement'':281,350,397,533,637,668,720,727,752,770 ''improv'':1012,1015 ''includ'':332 ''input'':684 ''int'':581,594 ''intefac'':388,730 ''interfac'':473,537,554,658,757 ''internation'':1022,1025 ''issu'':187 ''john'':349 ''json'':299,310,320 ''know'':1030 ''later'':871 ''light'':178 ''like'':542,612,631,688,779,854 ''limit'':344,579 ''list'':283,364,527 ''littl'':760 ''ll'':360,725,811,862 ''local'':1026,1028 ''logic'':400 ''look'':410,540 ''lot'':402,794 ''lucki'':799 ''m'':785 ''make'':376,456,499,698,764 ''mate'':629 ''may'':539 ''memori'':719 ''memory/cache'':736 ''menu'':164 ''merg'':86,111,114 ''method'':366,384,463,529,547,602,607,615,640,651,671,710,744,817,844,882,904,925 ''midway'':201 ''might'':789,1036 ''mkdir'':21 ''much'':494,682,1032 ''need'':137,369,618,654 ''new'':76,216 ''next'':63,714 ''note'':832 ''notic'':598 ''obsolet'':133 ''ofcours'':833 ''offset'':346,580 ''one'':358,391 ''open'':16,106,134,146,156 ''option'':335 ''order'':189 ''origin/feature-feed-service'':112 ''origin/next'':74 ''part'':395 ''passhash'':586 ''passhashiscorrect'':584,600 ''pattern'':348,575 ''perform'':1011,1016 ''pgadmin'':135 ''plan'':246,877 ''post'':293,954,965 ''postgr'':923 ''postgresql'':775 ''postid'':593 ''press'':1065 ''project'':12,192,259,451,502,701,767,1042,1049,1059 ''pull'':826 ''put'':311,957 ''queri'':157,173,177 ''question'':235,418,512 ''re'':278,428,830,835,1063 ''read'':194,205,227 ''realli'':444,483 ''reccommend'':484 ''rembrandt'':8C ''repositori'':648,657,677,709,723,729,756,773,903,924 ''request'':827 ''requir'':613 ''rest'':979 ''right'':818 ''run'':139,176,503,702,768 ''school'':6C ''search'':334,896,918,938 ''searchus'':574 ''secur'':993,997 ''sent'':212 ''sercvic'':121,806 ''servic'':73,83,105,276,371,387,462,472,520,536,544,549,553,643,662,847,859,881 ''setup'':13,875 ''shit'':975 ''simpl'':240,673,734 ''sinc'':441,798 ''slim'':23 ''slim-crown'':22 ''someth'':541 ''soon'':361 ''sortbi'':338,576 ''sortord'':342,577 ''special'':614,843 ''specif'':706 ''specifi'':382,545 ''sql'':132 ''start'':267,269 ''step'':783 ''stop'':200 ''string'':561,566,572,578,587,592 ''structur'':260 ''stuff'':687,808,970 ''succes'':207 ''sucki'':445 ''sure'':408,457,500,699,765 ''syllabus'':990 ''symbol'':179 ''test'':815,1001,1004 ''text'':2A,169,174,241 ''ther'':599 ''thing'':197,247,251,377,446 ''think'':623 ''time'':341,438,516,695 ''tip'':3,5,193 ''tips1.md'':10A ''tool'':52,158,167,182 ''toughest'':394 ''tri'':152,621 ''tricki'':399 ''type'':552 ''u'':567 ''understand'':256,425 ''unlucki'':38 ''updat'':890,912,933 ''update/create'':314 ''updateus'':564 ''url'':286 ''us'':1038 ''use'':238,466,713,733,774,872,1018,1024 ''user'':72,82,104,296,306,315,329,550,556,557,562,568,582,858,885,888,891,894,897,900,907,910,913,916,919,922,928,931,934,937,939,942 ''usernam'':304,308,313,317,323,560,565,571,585,591,949,959,963,967 ''usual'':143 ''valid'':685 ''ve'':705 ''way'':497,661 ''web'':994,1002,1005,1013 ''week'':991,999,1007,1020 ''whole'':196 ''wi'':44 ''wi-fi'':43 ''without'':226,644 ''work'':87,115,220,268,865,879,985 ''worri'':645 ''would'':434 ''x'':880,883,886,889,892,895,898,901,905,908,911,914,917,920,926,929,932,935,940,946,950,953,956,960,964 ''y'':424');
INSERT INTO "issue#1".tsvs_release VALUES (53, '''broken'':4 ''cartoon'':6C,10B ''chromagnum'':1B ''imag'':2A ''leav'':14A,18A ''moon'':5 ''pink'':15A ''pleas'':11A,19A ''rebbecca'':8C ''sugar'':9C');
INSERT INTO "issue#1".tsvs_release VALUES (52, '''27'':24 ''beckon'':18 ''call'':22 ''chromagnum'':1B ''clock'':26 ''edg'':17 ''end'':15A ''forget'':4 ''forget-this-not'':3 ''heed'':19 ''hood'':10C ''journey'':14A ''messag'':6C,12B ''messeng'':11C ''o'':25 ''omnious'':5C ''propheci'':8C ''text'':2A');
INSERT INTO "issue#1".tsvs_release VALUES (54, '''archaic'':12 ''blue'':36 ''chromagnum'':1B ''deep'':10 ''feather'':29 ''futur'':6 ''guidelin'':3 ''honest'':11 ''hood'':18C ''k'':15C ''k-pop'':14C ''literatur'':20B ''messeng'':19C ''path'':13 ''pop'':16C ''ruffl'':27 ''see'':34 ''sky'':37 ''text'':2A ''wind'':26 ''yet'':24A,30');
INSERT INTO "issue#1".tsvs_release VALUES (68, '''1'':14 ''39'':17 ''atom'':8B ''catastroph'':5C ''chromagnum'':1B ''full'':3 ''man'':7C ''midnight'':20 ''minut'':15 ''second'':18 ''stop'':4 ''test'':13A ''text'':2A');


--
-- Data for Name: user_avatars; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".user_avatars VALUES ('loveless', 'user.66cbf2be-c0ef-4056-8d43-aa7dca3e718b.lovelessness.jpg');


--
-- Data for Name: user_bookmarks; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 3, '2019-12-29 20:35:52.794761+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 4, '2020-01-12 10:12:37.060716+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 5, '2020-01-12 10:12:43.839661+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('rembrandtian', 5, '2019-12-29 20:28:43.700899+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('rembrandtian', 3, '2019-12-29 20:29:07.679261+03');


--
-- Data for Name: users; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".users VALUES ('dandelion@chroma.coma', 'haroma', '2020-01-26 19:42:52.97117+03', '$2a$10$68pXD8zRnV5hTj23nBX1Au3BpKeRXIrWQwUC0WI76kzCBf1PHUe86', 'Mag', NULL, 'Animal');
INSERT INTO "issue#1".users VALUES ('shuggie@shag.sho', 'faberge', '2020-01-26 20:10:16.561102+03', '$2a$10$1sdp1No9bGp3jgDoHa8foe67UUa.GuSmllvk6wjZ46V9Kqxji4PP6', 'We', 'Are', 'Destroyer');
INSERT INTO "issue#1".users VALUES ('yat@yayat.yat', 'slimmy', '2019-12-28 23:05:11.662742+03', '$2a$10$9tMmjrE2nv.KxvL0HIhk6.KmgwzPtX4MJ0YNSnEMNFjrERcJMjYRC', 'Kebab', 'Bab', 'Bob');
INSERT INTO "issue#1".users VALUES ('stars@destination.com', 'loveless', '2020-01-08 01:49:31.153913+03', '$2a$10$Oiok9PFU7iPW4O2mDl/2iOr.13LAxO1CrggiHleTP5r3N7gPxGyv6', 'Jeff', 'k.', 'Shoes');
INSERT INTO "issue#1".users VALUES ('serato@saskia.com', 'rembrandt', '2019-12-28 22:52:07.479752+03', '$2a$10$l689eOM7BwuHWqB9J0wM8OuGFiYXzJ0fo4UL1Dc2T.mRmLLvwys/e', 'Y.', 'A.', 'Knowe');
INSERT INTO "issue#1".users VALUES ('hot@hotter.hottest', 'rembrandtian', '2019-12-28 22:54:02.72085+03', '$2a$10$K14lpdifmeHPHxA6MptGKe10DlwbaXiKb4fNfmSUqG8hK5LeKWGhK', 'death', NULL, NULL);
INSERT INTO "issue#1".users VALUES ('zion@magnifico.com', 'IsisCane', '2020-01-16 21:44:42.22749+03', '$2a$10$99IjDKhuTCZ1qhCGOKIWC.fSS1HUFuCsfwGXrTwhSwqjkaj6.osWe', 'Devil', 'Laughter', 'Ramses');


--
-- Data for Name: users_bio; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".users_bio VALUES ('slimmy', 'get out');
INSERT INTO "issue#1".users_bio VALUES ('rembrandt', 'posh gorrila');
INSERT INTO "issue#1".users_bio VALUES ('loveless', 'i don&#39;t know what&#39;s real!');
INSERT INTO "issue#1".users_bio VALUES ('IsisCane', 'War on your nation and fire to your idols, you scum!');


--
-- Name: comments_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".comments_id_seq', 17, true);


--
-- Name: feeds_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".feeds_id_seq', 16, true);


--
-- Name: post_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".post_id_seq', 5, true);


--
-- Name: title_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".title_id_seq', 73, true);


--
-- Name: channel_admins channel_admins_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_pkey PRIMARY KEY (channel_username, username);


--
-- Name: channel_official_catalog channel_catalog_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_pkey PRIMARY KEY (channel_username, release_id);


--
-- Name: channel_pictures channel_pictures_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_pictures
    ADD CONSTRAINT channel_pictures_pk PRIMARY KEY (channelname);


--
-- Name: channels channels_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channels
    ADD CONSTRAINT channels_pkey PRIMARY KEY (username);


--
-- Name: tsvs_comment comment_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_comment
    ADD CONSTRAINT comment_tsvs_pk PRIMARY KEY (comment_id);


--
-- Name: comments comments_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comments
    ADD CONSTRAINT comments_pk PRIMARY KEY (id);


--
-- Name: feed_subscriptions feed_channels_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feed_subscriptions
    ADD CONSTRAINT feed_channels_pkey PRIMARY KEY (feed_id, channel_username);


--
-- Name: feeds feeds_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feeds
    ADD CONSTRAINT feeds_pkey PRIMARY KEY (id);


--
-- Name: releases_image_based image based_image_name_key; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_image_based
    ADD CONSTRAINT "image based_image_name_key" UNIQUE (image_name);


--
-- Name: releases_image_based image based_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_image_based
    ADD CONSTRAINT "image based_pkey" PRIMARY KEY (release_id);


--
-- Name: release_metadata metadata_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".release_metadata
    ADD CONSTRAINT metadata_pkey PRIMARY KEY (release_id);


--
-- Name: post_contents post_contents_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_contents
    ADD CONSTRAINT post_contents_pkey PRIMARY KEY (post_id, release_id);


--
-- Name: posts post_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- Name: post_stars post_stars_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_stars
    ADD CONSTRAINT post_stars_pkey PRIMARY KEY (username, post_id);


--
-- Name: tsvs_posts posts_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_posts
    ADD CONSTRAINT posts_tsvs_pk PRIMARY KEY (post_id);


--
-- Name: releases release_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases
    ADD CONSTRAINT release_pkey PRIMARY KEY (id);


--
-- Name: tsvs_release release_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_release
    ADD CONSTRAINT release_tsvs_pk PRIMARY KEY (release_id);


--
-- Name: releases_text_based text based_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_text_based
    ADD CONSTRAINT "text based_pkey" PRIMARY KEY (release_id);


--
-- Name: user_avatars user_avatars_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_avatars
    ADD CONSTRAINT user_avatars_pk PRIMARY KEY (username);


--
-- Name: user_bookmarks user_bookmarks_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_bookmarks
    ADD CONSTRAINT user_bookmarks_pkey PRIMARY KEY (username, post_id);


--
-- Name: users_bio users_bio_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users_bio
    ADD CONSTRAINT users_bio_pk PRIMARY KEY (username);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users
    ADD CONSTRAINT users_pkey PRIMARY KEY (username);


--
-- Name: comment_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX comment_ts_index ON "issue#1".tsvs_comment USING gin (vector);


--
-- Name: post_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX post_ts_index ON "issue#1".tsvs_posts USING gin (vector);


--
-- Name: release_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX release_ts_index ON "issue#1".tsvs_release USING gin (vector);


--
-- Name: release_tsvs_release_id_uindex; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE UNIQUE INDEX release_tsvs_release_id_uindex ON "issue#1".tsvs_release USING btree (release_id);


--
-- Name: comments comment_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER comment_insert_trigger AFTER INSERT ON "issue#1".comments FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_comment_trigger();


--
-- Name: comments comment_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER comment_update_trigger AFTER UPDATE ON "issue#1".comments FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_comment_trigger();


--
-- Name: release_metadata metadata_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER metadata_insert_trigger AFTER INSERT ON "issue#1".release_metadata FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_on_metadata_update_trigger();


--
-- Name: release_metadata metadata_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER metadata_update_trigger AFTER UPDATE ON "issue#1".release_metadata FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_on_metadata_update_trigger();


--
-- Name: posts post_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER post_insert_trigger AFTER INSERT ON "issue#1".posts FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_post_trigger();


--
-- Name: posts post_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER post_update_trigger AFTER UPDATE ON "issue#1".posts FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_post_trigger();


--
-- Name: users setup_user; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER setup_user AFTER INSERT ON "issue#1".users FOR EACH ROW EXECUTE FUNCTION "issue#1".setup_user();


--
-- Name: releases_text_based text_based_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER text_based_update_trigger AFTER UPDATE ON "issue#1".releases_text_based FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_text_based_update_trigger();


--
-- Name: channel_admins channel_admins_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_channel_username_fkey FOREIGN KEY (channel_username) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: channel_admins channel_admins_user_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_user_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: channel_official_catalog channel_catalog_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_channel_username_fkey FOREIGN KEY (channel_username) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: channel_official_catalog channel_catalog_post_from_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_post_from_id_fkey FOREIGN KEY (post_from_id) REFERENCES "issue#1".posts(id);


--
-- Name: channel_official_catalog channel_catalog_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: channel_pictures channel_pictures_channels_username_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_pictures
    ADD CONSTRAINT channel_pictures_channels_username_fk FOREIGN KEY (channelname) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tsvs_comment comment_tsvs_comments_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_comment
    ADD CONSTRAINT comment_tsvs_comments_id_fk FOREIGN KEY (comment_id) REFERENCES "issue#1".comments(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: comments comments_commented_by_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comments
    ADD CONSTRAINT comments_commented_by_fkey FOREIGN KEY (commented_by) REFERENCES "issue#1".users(username) ON UPDATE CASCADE NOT VALID;


--
-- Name: comments comments_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comments
    ADD CONSTRAINT comments_post_id_fkey FOREIGN KEY (post_from) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: feed_subscriptions feed_channels_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feed_subscriptions
    ADD CONSTRAINT feed_channels_channel_username_fkey FOREIGN KEY (channel_username) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: feed_subscriptions feed_channels_feed_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feed_subscriptions
    ADD CONSTRAINT feed_channels_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES "issue#1".feeds(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: feeds feeds_owner_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feeds
    ADD CONSTRAINT feeds_owner_username_fkey FOREIGN KEY (owner_username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: release_metadata metadata_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".release_metadata
    ADD CONSTRAINT metadata_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON DELETE CASCADE;


--
-- Name: post_contents post_contents_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_contents
    ADD CONSTRAINT post_contents_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: post_contents post_contents_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_contents
    ADD CONSTRAINT post_contents_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: post_stars post_stars_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_stars
    ADD CONSTRAINT post_stars_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: post_stars post_stars_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_stars
    ADD CONSTRAINT post_stars_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: posts posts_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts
    ADD CONSTRAINT posts_channel_username_fkey FOREIGN KEY (channel_from) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: posts posts_poster_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts
    ADD CONSTRAINT posts_poster_username_fkey FOREIGN KEY (posted_by) REFERENCES "issue#1".users(username) ON UPDATE CASCADE NOT VALID;


--
-- Name: tsvs_posts posts_tsvs_posts_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_posts
    ADD CONSTRAINT posts_tsvs_posts_id_fk FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tsvs_release release_tsvs_releases_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_release
    ADD CONSTRAINT release_tsvs_releases_id_fk FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: releases_image_based releases_image_based_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_image_based
    ADD CONSTRAINT releases_image_based_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: releases releases_owner_channel_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases
    ADD CONSTRAINT releases_owner_channel_fkey FOREIGN KEY (owner_channel) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: channel_stickies stickied_posts_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_stickies
    ADD CONSTRAINT stickied_posts_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: releases_text_based text based_title_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_text_based
    ADD CONSTRAINT "text based_title_id_fkey" FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON DELETE CASCADE NOT VALID;


--
-- Name: user_avatars user_avatars_users_username_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_avatars
    ADD CONSTRAINT user_avatars_users_username_fk FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_bookmarks user_bookmarks_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_bookmarks
    ADD CONSTRAINT user_bookmarks_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_bookmarks user_bookmarks_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_bookmarks
    ADD CONSTRAINT user_bookmarks_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: users_bio users_bio_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users_bio
    ADD CONSTRAINT users_bio_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: FUNCTION citextin(cstring); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextin(cstring) TO "issue#1_REST";


--
-- Name: FUNCTION citextout("issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextout("issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citextrecv(internal); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextrecv(internal) TO "issue#1_REST";


--
-- Name: FUNCTION citextsend("issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextsend("issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext(boolean); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext(boolean) TO "issue#1_REST";


--
-- Name: FUNCTION citext(character); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext(character) TO "issue#1_REST";


--
-- Name: FUNCTION citext(inet); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext(inet) TO "issue#1_REST";


--
-- Name: FUNCTION citext_cmp("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_cmp("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_eq("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_eq("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_ge("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_ge("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_gt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_gt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_hash("issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_hash("issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_hash_extended("issue#1".citext, bigint); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_hash_extended("issue#1".citext, bigint) TO "issue#1_REST";


--
-- Name: FUNCTION citext_larger("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_larger("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_le("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_le("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_lt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_lt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_ne("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_ne("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_pattern_cmp("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_cmp("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_pattern_ge("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_ge("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_pattern_gt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_gt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_pattern_le("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_le("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_pattern_lt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_lt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION citext_smaller("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_smaller("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_match("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_match("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_match("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_match("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_matches("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_matches("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_matches("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_matches("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_replace("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_replace("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_replace("issue#1".citext, "issue#1".citext, text, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_replace("issue#1".citext, "issue#1".citext, text, text) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_split_to_array("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_array("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_split_to_array("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_array("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_split_to_table("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_table("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION regexp_split_to_table("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_table("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION replace("issue#1".citext, "issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".replace("issue#1".citext, "issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION split_part("issue#1".citext, "issue#1".citext, integer); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".split_part("issue#1".citext, "issue#1".citext, integer) TO "issue#1_REST";


--
-- Name: FUNCTION strpos("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".strpos("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION texticlike("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticlike("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION texticlike("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticlike("issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION texticnlike("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticnlike("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION texticnlike("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticnlike("issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION texticregexeq("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexeq("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION texticregexeq("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexeq("issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION texticregexne("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexne("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- Name: FUNCTION texticregexne("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexne("issue#1".citext, text) TO "issue#1_REST";


--
-- Name: FUNCTION translate("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".translate("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- Name: TABLE channel_admins; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channel_admins TO "issue#1_REST";


--
-- Name: TABLE channel_official_catalog; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channel_official_catalog TO "issue#1_REST";


--
-- Name: TABLE channel_stickies; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channel_stickies TO "issue#1_REST";


--
-- Name: TABLE channels; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channels TO "issue#1_REST";


--
-- Name: TABLE comments; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".comments TO "issue#1_REST";


--
-- Name: TABLE feed_subscriptions; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".feed_subscriptions TO "issue#1_REST";


--
-- Name: TABLE feeds; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".feeds TO "issue#1_REST";


--
-- Name: SEQUENCE feeds_id_seq; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON SEQUENCE "issue#1".feeds_id_seq TO "issue#1_REST";


--
-- Name: TABLE post_contents; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".post_contents TO "issue#1_REST";


--
-- Name: TABLE posts; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".posts TO "issue#1_REST";


--
-- Name: SEQUENCE post_id_seq; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON SEQUENCE "issue#1".post_id_seq TO "issue#1_REST";


--
-- Name: TABLE post_stars; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".post_stars TO "issue#1_REST";


--
-- Name: TABLE release_metadata; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".release_metadata TO "issue#1_REST";


--
-- Name: TABLE releases; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".releases TO "issue#1_REST";


--
-- Name: TABLE releases_image_based; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".releases_image_based TO "issue#1_REST";


--
-- Name: TABLE releases_text_based; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".releases_text_based TO "issue#1_REST";


--
-- Name: SEQUENCE title_id_seq; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON SEQUENCE "issue#1".title_id_seq TO "issue#1_REST";


--
-- Name: TABLE user_bookmarks; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".user_bookmarks TO "issue#1_REST";


--
-- Name: TABLE users; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".users TO "issue#1_REST";


--
-- Name: TABLE users_bio; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".users_bio TO "issue#1_REST";


--
-- PostgreSQL database dump complete
--

