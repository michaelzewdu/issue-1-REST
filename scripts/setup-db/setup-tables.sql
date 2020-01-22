
--
-- TOC entry 2 (class 3079 OID 98833)
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA "issue#1";


--
-- TOC entry 3145 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- TOC entry 288 (class 1255 OID 123309)
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
-- TOC entry 289 (class 1255 OID 123420)
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
-- TOC entry 286 (class 1255 OID 123330)
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
-- TOC entry 287 (class 1255 OID 123357)
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
-- TOC entry 290 (class 1255 OID 123338)
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
-- TOC entry 203 (class 1259 OID 98938)
-- Name: channel_admins; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_admins (
    channel_username character varying(24) NOT NULL,
    username character varying(24) NOT NULL,
    is_owner boolean NOT NULL
);


ALTER TABLE "issue#1".channel_admins OWNER TO "issue#1_dev";

--
-- TOC entry 204 (class 1259 OID 98941)
-- Name: channel_official_catalog; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_official_catalog (
    channel_username character varying(24) NOT NULL,
    release_id integer NOT NULL,
    post_from_id integer NOT NULL
);


ALTER TABLE "issue#1".channel_official_catalog OWNER TO "issue#1_dev";

--
-- TOC entry 224 (class 1259 OID 115130)
-- Name: channel_pictures; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_pictures (
    channelname character varying(24) NOT NULL,
    image_name text NOT NULL
);


ALTER TABLE "issue#1".channel_pictures OWNER TO "issue#1_dev";

--
-- TOC entry 205 (class 1259 OID 98944)
-- Name: channel_stickies; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_stickies (
    post_id integer
);


ALTER TABLE "issue#1".channel_stickies OWNER TO "issue#1_dev";

--
-- TOC entry 206 (class 1259 OID 98947)
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
-- TOC entry 207 (class 1259 OID 98954)
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
-- TOC entry 225 (class 1259 OID 123311)
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
-- TOC entry 208 (class 1259 OID 98961)
-- Name: feed_subscriptions; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".feed_subscriptions (
    feed_id integer NOT NULL,
    channel_username character varying(24) NOT NULL,
    subscription_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "issue#1".feed_subscriptions OWNER TO "issue#1_dev";

--
-- TOC entry 209 (class 1259 OID 98965)
-- Name: feeds; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".feeds (
    owner_username character varying(24) NOT NULL,
    sorting text NOT NULL,
    id integer NOT NULL
);


ALTER TABLE "issue#1".feeds OWNER TO "issue#1_dev";

--
-- TOC entry 210 (class 1259 OID 98971)
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
-- TOC entry 3198 (class 0 OID 0)
-- Dependencies: 210
-- Name: feeds_id_seq; Type: SEQUENCE OWNED BY; Schema: issue#1; Owner: issue#1_dev
--

ALTER SEQUENCE "issue#1".feeds_id_seq OWNED BY "issue#1".feeds.id;


--
-- TOC entry 213 (class 1259 OID 98985)
-- Name: post_contents; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".post_contents (
    post_id integer NOT NULL,
    release_id integer NOT NULL
);


ALTER TABLE "issue#1".post_contents OWNER TO "issue#1_dev";

--
-- TOC entry 214 (class 1259 OID 98988)
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
-- TOC entry 215 (class 1259 OID 98995)
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
-- TOC entry 3202 (class 0 OID 0)
-- Dependencies: 215
-- Name: post_id_seq; Type: SEQUENCE OWNED BY; Schema: issue#1; Owner: issue#1_dev
--

ALTER SEQUENCE "issue#1".post_id_seq OWNED BY "issue#1".posts.id;


--
-- TOC entry 216 (class 1259 OID 98997)
-- Name: post_stars; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".post_stars (
    star_count integer NOT NULL,
    post_id integer NOT NULL,
    username character varying(22) NOT NULL
);


ALTER TABLE "issue#1".post_stars OWNER TO "issue#1_dev";

--
-- TOC entry 212 (class 1259 OID 98979)
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
-- TOC entry 217 (class 1259 OID 99000)
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
-- TOC entry 211 (class 1259 OID 98973)
-- Name: releases_image_based; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".releases_image_based (
    release_id integer NOT NULL,
    image_name text NOT NULL
);


ALTER TABLE "issue#1".releases_image_based OWNER TO "issue#1_dev";

--
-- TOC entry 218 (class 1259 OID 99004)
-- Name: releases_text_based; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".releases_text_based (
    release_id integer NOT NULL,
    content text NOT NULL
);


ALTER TABLE "issue#1".releases_text_based OWNER TO "issue#1_dev";

--
-- TOC entry 219 (class 1259 OID 99010)
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
-- TOC entry 3209 (class 0 OID 0)
-- Dependencies: 219
-- Name: title_id_seq; Type: SEQUENCE OWNED BY; Schema: issue#1; Owner: issue#1_dev
--

ALTER SEQUENCE "issue#1".title_id_seq OWNED BY "issue#1".releases.id;


--
-- TOC entry 228 (class 1259 OID 123406)
-- Name: tsvs_comment; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".tsvs_comment (
    comment_id integer NOT NULL,
    vector tsvector
);


ALTER TABLE "issue#1".tsvs_comment OWNER TO "issue#1_dev";

--
-- TOC entry 227 (class 1259 OID 123344)
-- Name: tsvs_posts; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".tsvs_posts (
    post_id integer NOT NULL,
    vector tsvector
);


ALTER TABLE "issue#1".tsvs_posts OWNER TO "issue#1_dev";

--
-- TOC entry 226 (class 1259 OID 123320)
-- Name: tsvs_release; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".tsvs_release (
    release_id integer NOT NULL,
    vector tsvector
);


ALTER TABLE "issue#1".tsvs_release OWNER TO "issue#1_dev";

--
-- TOC entry 223 (class 1259 OID 115117)
-- Name: user_avatars; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".user_avatars (
    username character varying(24) NOT NULL,
    image_name text NOT NULL
);


ALTER TABLE "issue#1".user_avatars OWNER TO "issue#1_dev";

--
-- TOC entry 220 (class 1259 OID 99012)
-- Name: user_bookmarks; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".user_bookmarks (
    username character varying(24) NOT NULL,
    post_id integer NOT NULL,
    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "issue#1".user_bookmarks OWNER TO "issue#1_dev";

--
-- TOC entry 221 (class 1259 OID 99016)
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
-- TOC entry 222 (class 1259 OID 99023)
-- Name: users_bio; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".users_bio (
    username character varying(24) NOT NULL,
    bio text NOT NULL
);


ALTER TABLE "issue#1".users_bio OWNER TO "issue#1_dev";

--
-- TOC entry 2891 (class 2604 OID 99029)
-- Name: feeds id; Type: DEFAULT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feeds ALTER COLUMN id SET DEFAULT nextval('"issue#1".feeds_id_seq'::regclass);


--
-- TOC entry 2894 (class 2604 OID 99030)
-- Name: posts id; Type: DEFAULT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts ALTER COLUMN id SET DEFAULT nextval('"issue#1".post_id_seq'::regclass);


--
-- TOC entry 2895 (class 2604 OID 99031)
-- Name: releases id; Type: DEFAULT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases ALTER COLUMN id SET DEFAULT nextval('"issue#1".title_id_seq'::regclass);


--
-- TOC entry 3112 (class 0 OID 98938)
-- Dependencies: 203
-- Data for Name: channel_admins; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".channel_admins VALUES ('Isis Cane', 'Isis Cane', true);


--
-- TOC entry 3113 (class 0 OID 98941)
-- Dependencies: 204
-- Data for Name: channel_official_catalog; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".channel_official_catalog VALUES ('chromagnum', 6, 4);


--
-- TOC entry 3133 (class 0 OID 115130)
-- Dependencies: 224
-- Data for Name: channel_pictures; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- TOC entry 3114 (class 0 OID 98944)
-- Dependencies: 205
-- Data for Name: channel_stickies; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- TOC entry 3115 (class 0 OID 98947)
-- Dependencies: 206
-- Data for Name: channels; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".channels VALUES ('chromagnum', '2019-12-29 19:59:22.564274+03', 'THE FUTURE IS NOW', 'take it off');
INSERT INTO "issue#1".channels VALUES ('Isis Cane', '2020-01-16 21:44:42.22749+03', 'Isis Cane''s channel', NULL);
INSERT INTO "issue#1".channels VALUES ('tempOne', '2020-01-19 23:49:54.013003+03', 'tempOne''s channel', NULL);


--
-- TOC entry 3116 (class 0 OID 98954)
-- Dependencies: 207
-- Data for Name: comments; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 3, -1, '<p>they can hear me crying… </p>
', 'loveless', '2020-01-11 20:16:33.519527+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 2, 1, 'duhh', 'rembrandtian', '2020-01-11 19:30:04.273774+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 6, 3, '<p>the diamond mines are <em>burning</em> </p>
', 'rembrandtian', '2020-01-11 20:26:34.222721+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 8, -1, 'It''s easy if you''re passionate...', 'loveless', '2020-01-15 02:14:45.403771+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 1, -1, 'Epistien did not kill himself!', 'slimmy', '2020-01-11 19:29:29.058057+03');


--
-- TOC entry 3117 (class 0 OID 98961)
-- Dependencies: 208
-- Data for Name: feed_subscriptions; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".feed_subscriptions VALUES (3, 'chromagnum', '2020-01-06 17:01:08.862048+03');


--
-- TOC entry 3118 (class 0 OID 98965)
-- Dependencies: 209
-- Data for Name: feeds; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".feeds VALUES ('rembrandt', 'hot', 3);
INSERT INTO "issue#1".feeds VALUES ('rembrandtian', 'top', 6);
INSERT INTO "issue#1".feeds VALUES ('slimmy', 'top', 7);
INSERT INTO "issue#1".feeds VALUES ('loveless', 'new', 5);
INSERT INTO "issue#1".feeds VALUES ('Isis Cane', 'hot', 8);


--
-- TOC entry 3122 (class 0 OID 98985)
-- Dependencies: 213
-- Data for Name: post_contents; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- TOC entry 3125 (class 0 OID 98997)
-- Dependencies: 216
-- Data for Name: post_stars; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- TOC entry 3123 (class 0 OID 98988)
-- Dependencies: 214
-- Data for Name: posts; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".posts VALUES (4, 'tipstitipst[tipstst]tipstitstipstsi[ps]', 'HEREWEGOLANG', 'rembrandt', 'chromagnum', '2020-01-04 20:58:27.926859+03');
INSERT INTO "issue#1".posts VALUES (5, NULL, 'cordovadovadovadovedova', 'rembrandt', 'chromagnum', '2019-12-29 20:00:53.714587+03');
INSERT INTO "issue#1".posts VALUES (3, 'B7 Chord, G6 Chord.', 'its so strangeeee', 'loveless', 'chromagnum', '2019-12-29 19:59:39.568527+03');
INSERT INTO "issue#1".posts VALUES (6, 'Welcome!', 'Issue #1 v0.1', 'rembrandt', 'chromagnum', '2020-01-12 18:54:48.460537+03');


--
-- TOC entry 3121 (class 0 OID 98979)
-- Dependencies: 212
-- Data for Name: release_metadata; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".release_metadata VALUES (6, 'tips baby, tips', '{"genres": ["School"], "authors": ["@rembrandt"]}', 'Guide', '2020-01-04 11:25:15.055178+03', 'Tips1.md');
INSERT INTO "issue#1".release_metadata VALUES (53, 'a broken moon', '{"genres": ["Cartoon"], "authors": ["Rebbecca Sugar"]}', 'Cartoon', '2016-01-04 23:06:16.017584+03', 'Please don''t leave Pink...don''t leave please.');
INSERT INTO "issue#1".release_metadata VALUES (52, 'Forget-this-not.', '{"genres": ["Omnious Message", "Prophecy"], "authors": ["Hooded Messenger"]}', 'Message', '2020-01-05 13:14:16.017584+03', 'The Journey Ends!');
INSERT INTO "issue#1".release_metadata VALUES (54, 'Guidelines to the future and to the deep, honest, archaic path.', '{"genres": ["K-Pop"], "authors": ["Hooded Messenger"]}', 'Literature', '2020-12-05 13:14:16.017584+03', 'Above & Not There Yet.');


--
-- TOC entry 3126 (class 0 OID 99000)
-- Dependencies: 217
-- Data for Name: releases; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".releases VALUES (6, 'chromagnum', 'text', '2020-01-04 11:23:04.271553+03');
INSERT INTO "issue#1".releases VALUES (52, 'chromagnum', 'text', '2020-01-05 13:18:43.7471+03');
INSERT INTO "issue#1".releases VALUES (53, 'chromagnum', 'image', '2020-01-05 13:55:05.730832+03');
INSERT INTO "issue#1".releases VALUES (54, 'chromagnum', 'text', '2020-01-12 18:05:22.383129+03');


--
-- TOC entry 3120 (class 0 OID 98973)
-- Dependencies: 211
-- Data for Name: releases_image_based; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".releases_image_based VALUES (53, 'release.1eb2969c-cf4c-40cb-9b78-39bea2b089da.the image.jpg');


--
-- TOC entry 3127 (class 0 OID 99004)
-- Dependencies: 218
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


--
-- TOC entry 3137 (class 0 OID 123406)
-- Dependencies: 228
-- Data for Name: tsvs_comment; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".tsvs_comment VALUES (3, '''cri'':5 ''hear'':3 ''loveless'':6B');
INSERT INTO "issue#1".tsvs_comment VALUES (2, '''duhh'':1 ''rembrandtian'':2B');
INSERT INTO "issue#1".tsvs_comment VALUES (6, '''burn'':5 ''diamond'':2 ''mine'':3 ''rembrandtian'':6B');
INSERT INTO "issue#1".tsvs_comment VALUES (8, '''easi'':3 ''loveless'':8B ''passion'':7 ''re'':6');
INSERT INTO "issue#1".tsvs_comment VALUES (1, '''epistien'':1 ''kill'':4 ''slimmi'':5B');


--
-- TOC entry 3136 (class 0 OID 123344)
-- Dependencies: 227
-- Data for Name: tsvs_posts; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".tsvs_posts VALUES (4, '''chromagnum'':7B ''herewegolang'':5A ''ps'':4 ''rembrandt'':6B ''tipstitipst'':1 ''tipstitstipstsi'':3 ''tipstst'':2');
INSERT INTO "issue#1".tsvs_posts VALUES (5, '''chromagnum'':3B ''cordovadovadovadovedova'':1A ''rembrandt'':2B');
INSERT INTO "issue#1".tsvs_posts VALUES (3, '''b7'':1 ''chord'':2,4 ''chromagnum'':9B ''g6'':3 ''loveless'':8B ''strangeee'':7A');
INSERT INTO "issue#1".tsvs_posts VALUES (6, '''1'':3A ''chromagnum'':6B ''issu'':2A ''rembrandt'':5B ''v0.1'':4A ''welcom'':1');


--
-- TOC entry 3135 (class 0 OID 123320)
-- Dependencies: 226
-- Data for Name: tsvs_release; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".tsvs_release VALUES (6, '''/go/src/github.com'':20 ''/go/src/github.com/slim-crown/issue-1-rest'':66,99 ''/gorilla/mux'':131 ''/slim-crown/issue-1-rest.git'':29 ''/slim-crown/issue-1-website.git'':34 ''/users'':294,303,312,322,337,948,952,955,958,962,966 ''0'':347 ''1.sql'':188 ''10'':1054 ''11'':992 ''12'':1000 ''13'':1008 ''14'':1021 ''20'':1046 ''25'':345 ''abl'':813 ''accord'':987 ''add'':996 ''addus'':555 ''administr'':141 ''afraid'':786 ''ahead'':225,248 ''along'':659 ''also'':237,331,692,739,1041 ''api'':980 ''applic'':995,1003,1006,1014 ''area'':175 ''articl'':210,229 ''asc'':343 ''ask'':48,230,417,507,1037 ''attempt'':150 ''authent'':855,899,968 ''author'':921,941 ''b'':69 ''babi'':4 ''bar'':165,183 ''base'':751 ''bash'':18 ''basic'':255,605 ''beleiv'':492 ''besid'':603,683 ''big'':1058 ''bookmarkpost'':590,601 ''bool'':588 ''boom'':828 ''branch'':60,64,77,91,97,108,122,209,217 ''cach'':761,902 ''call'':79,674,747 ''cd'':19,65,98 ''checkout'':68,101 ''chromagnum'':1B ''clone'':15,26,31 ''commit'':440,449,506,697,769 ''compar'':1047 ''compil'':455 ''complet'':837 ''concurr'':1009,1019 ''confid'':264 ''construct'':421 ''copi'':168 ''creat'':57,214,295,824 ''creation'':340 ''creation-tim'':339 ''crown'':24 ''crud'':606 ''db'':750 ''deadlin'':1073 ''defin'':459,650 ''delet'':321,893,915,936,961 ''deleteus'':570 ''deliv'':1067 ''depend'':126 ''design'':244 ''differ'':842 ''direct'':746 ''discuss'':627 ''doesn'':453 ''done'':88,116,429,486,831,838,972 ''e.g'':292,853 ''earler'':732 ''easi'':378,741 ''easier'':495 ''easiest'':666 ''end'':790 ''endpoint'':273 ''enough'':39,265 ''entit'':609 ''entiti'':551 ''error'':404,478,558,563,569,573,583,589,595,796 ''especi'':325 ''essenti'':191 ''evalu'':1043,1050 ''expect'':289 ''fail'':144 ''fast'':976 ''featur'':71,81,103,998,1069 ''feature-user-servic'':70,80,102 ''feed'':120 ''feed-sercvic'':119 ''feel'':263 ''fi'':45 ''file'':242 ''first'':109,149,202,353,809,1072 ''fix'':475 ''follow'':874 ''form'':301 ''found'':92,161 ''function'':413,620,945 ''get'':128,302,305,326,336,485,884,887,906,909,927,930,947,951,971 ''getus'':559 ''git'':14,17,25,30,67,100,110,208 ''github'':94 ''github.com'':28,33,130 ''github.com/gorilla/mux'':129 ''github.com/slim-crown/issue-1-rest.git'':27 ''github.com/slim-crown/issue-1-website.git'':32 ''given'':298,319 ''globals.sql'':171 ''go'':127,224,279,357,379,716 ''good'':437,694 ''guid'':9B ''guidelin'':11 ''half'':983 ''handl'':291,405,851 ''handler'':272,352,389,432,469,489,636,803,944 ''hard'':1064 ''hardest'':782 ''help'':54 ''home'':47 ''http'':943 ''ii'':1051 ''implememt'':1040 ''implement'':281,350,397,533,637,668,720,727,752,770 ''improv'':1012,1015 ''includ'':332 ''input'':684 ''int'':581,594 ''intefac'':388,730 ''interfac'':473,537,554,658,757 ''internation'':1022,1025 ''issu'':187 ''john'':349 ''json'':299,310,320 ''know'':1030 ''later'':871 ''light'':178 ''like'':542,612,631,688,779,854 ''limit'':344,579 ''list'':283,364,527 ''littl'':760 ''ll'':360,725,811,862 ''local'':1026,1028 ''logic'':400 ''look'':410,540 ''lot'':402,794 ''lucki'':799 ''m'':785 ''make'':376,456,499,698,764 ''mate'':629 ''may'':539 ''memori'':719 ''memory/cache'':736 ''menu'':164 ''merg'':86,111,114 ''method'':366,384,463,529,547,602,607,615,640,651,671,710,744,817,844,882,904,925 ''midway'':201 ''might'':789,1036 ''mkdir'':21 ''much'':494,682,1032 ''need'':137,369,618,654 ''new'':76,216 ''next'':63,714 ''note'':832 ''notic'':598 ''obsolet'':133 ''ofcours'':833 ''offset'':346,580 ''one'':358,391 ''open'':16,106,134,146,156 ''option'':335 ''order'':189 ''origin/feature-feed-service'':112 ''origin/next'':74 ''part'':395 ''passhash'':586 ''passhashiscorrect'':584,600 ''pattern'':348,575 ''perform'':1011,1016 ''pgadmin'':135 ''plan'':246,877 ''post'':293,954,965 ''postgr'':923 ''postgresql'':775 ''postid'':593 ''press'':1065 ''project'':12,192,259,451,502,701,767,1042,1049,1059 ''pull'':826 ''put'':311,957 ''queri'':157,173,177 ''question'':235,418,512 ''re'':278,428,830,835,1063 ''read'':194,205,227 ''realli'':444,483 ''reccommend'':484 ''rembrandt'':8C ''repositori'':648,657,677,709,723,729,756,773,903,924 ''request'':827 ''requir'':613 ''rest'':979 ''right'':818 ''run'':139,176,503,702,768 ''school'':6C ''search'':334,896,918,938 ''searchus'':574 ''secur'':993,997 ''sent'':212 ''sercvic'':121,806 ''servic'':73,83,105,276,371,387,462,472,520,536,544,549,553,643,662,847,859,881 ''setup'':13,875 ''shit'':975 ''simpl'':240,673,734 ''sinc'':441,798 ''slim'':23 ''slim-crown'':22 ''someth'':541 ''soon'':361 ''sortbi'':338,576 ''sortord'':342,577 ''special'':614,843 ''specif'':706 ''specifi'':382,545 ''sql'':132 ''start'':267,269 ''step'':783 ''stop'':200 ''string'':561,566,572,578,587,592 ''structur'':260 ''stuff'':687,808,970 ''succes'':207 ''sucki'':445 ''sure'':408,457,500,699,765 ''syllabus'':990 ''symbol'':179 ''test'':815,1001,1004 ''text'':2A,169,174,241 ''ther'':599 ''thing'':197,247,251,377,446 ''think'':623 ''time'':341,438,516,695 ''tip'':3,5,193 ''tips1.md'':10A ''tool'':52,158,167,182 ''toughest'':394 ''tri'':152,621 ''tricki'':399 ''type'':552 ''u'':567 ''understand'':256,425 ''unlucki'':38 ''updat'':890,912,933 ''update/create'':314 ''updateus'':564 ''url'':286 ''us'':1038 ''use'':238,466,713,733,774,872,1018,1024 ''user'':72,82,104,296,306,315,329,550,556,557,562,568,582,858,885,888,891,894,897,900,907,910,913,916,919,922,928,931,934,937,939,942 ''usernam'':304,308,313,317,323,560,565,571,585,591,949,959,963,967 ''usual'':143 ''valid'':685 ''ve'':705 ''way'':497,661 ''web'':994,1002,1005,1013 ''week'':991,999,1007,1020 ''whole'':196 ''wi'':44 ''wi-fi'':43 ''without'':226,644 ''work'':87,115,220,268,865,879,985 ''worri'':645 ''would'':434 ''x'':880,883,886,889,892,895,898,901,905,908,911,914,917,920,926,929,932,935,940,946,950,953,956,960,964 ''y'':424');
INSERT INTO "issue#1".tsvs_release VALUES (53, '''broken'':4 ''cartoon'':6C,10B ''chromagnum'':1B ''imag'':2A ''leav'':14A,18A ''moon'':5 ''pink'':15A ''pleas'':11A,19A ''rebbecca'':8C ''sugar'':9C');
INSERT INTO "issue#1".tsvs_release VALUES (52, '''27'':24 ''beckon'':18 ''call'':22 ''chromagnum'':1B ''clock'':26 ''edg'':17 ''end'':15A ''forget'':4 ''forget-this-not'':3 ''heed'':19 ''hood'':10C ''journey'':14A ''messag'':6C,12B ''messeng'':11C ''o'':25 ''omnious'':5C ''propheci'':8C ''text'':2A');
INSERT INTO "issue#1".tsvs_release VALUES (54, '''archaic'':12 ''blue'':36 ''chromagnum'':1B ''deep'':10 ''feather'':29 ''futur'':6 ''guidelin'':3 ''honest'':11 ''hood'':18C ''k'':15C ''k-pop'':14C ''literatur'':20B ''messeng'':19C ''path'':13 ''pop'':16C ''ruffl'':27 ''see'':34 ''sky'':37 ''text'':2A ''wind'':26 ''yet'':24A,30');


--
-- TOC entry 3132 (class 0 OID 115117)
-- Dependencies: 223
-- Data for Name: user_avatars; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".user_avatars VALUES ('loveless', 'user.66cbf2be-c0ef-4056-8d43-aa7dca3e718b.lovelessness.jpg');


--
-- TOC entry 3129 (class 0 OID 99012)
-- Dependencies: 220
-- Data for Name: user_bookmarks; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".user_bookmarks VALUES ('rembrandtian', 5, '2019-12-29 20:28:43.700899+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('rembrandtian', 3, '2019-12-29 20:29:07.679261+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 3, '2019-12-29 20:35:52.794761+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 4, '2020-01-12 10:12:37.060716+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 5, '2020-01-12 10:12:43.839661+03');


--
-- TOC entry 3130 (class 0 OID 99016)
-- Dependencies: 221
-- Data for Name: users; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".users VALUES ('hot@hotter.hottest', 'rembrandtian', '2019-12-28 22:54:02.72085+03', '\xafe6d7d32ac4fff5eba8f09debdaaf57f946bd12239409f0f4b161558c7988f17a7cc47bc099bf9afa0f951458609055acd0149855f0ac7b604ecbeb28a63f5d', 'death', NULL, NULL);
INSERT INTO "issue#1".users VALUES ('serato@saskia.com', 'rembrandt', '2019-12-28 22:52:07.479752+03', '\xe20d91786b7a9317bb0c84c11af36f3f459adb1c386347aab05b68e9eac8dd287689ef1924dd80309c67d8656a9bf520cc4c8dea7fe0196d675437d21eb1da20', 'Y.', 'A.', 'Knowe');
INSERT INTO "issue#1".users VALUES ('stars@destination.com', 'loveless', '2020-01-08 01:49:31.153913+03', '\x87bba9f7d30eb9f64bc2d77eace67147c370e7073d22cb5bedfc5ae8b5bd3ca60b1748e21df9626749c031208e9811e38c3204a7efdf8a867ebfb9b066325251', 'Jeff', 'k.', 'Shoes');
INSERT INTO "issue#1".users VALUES ('yat@yayat.yat', 'slimmy', '2019-12-28 23:05:11.662742+03', '\x58df5e6ebc6c6a8b4760f514652d7baaa864cc1e7e656a5734a253ad3b16c4a69561a202f3ae79bab9cbadc6ed20224b90ad7d467a3bfb773964a80e27d03e97', 'Kebab', 'Bab', 'Bob');
INSERT INTO "issue#1".users VALUES ('zion@magnifico.com', 'Isis Cane', '2020-01-16 21:44:42.22749+03', '\x0e5cbfc12ef5bf26041b1b2c953dc0f733a2717b642de99993e143753397b065b02ef65b6ada58b85376e7da2fd62772ada782489f48a69e27f5af3eb85e70f7', 'Devil', 'Laughter', 'Ramses');


--
-- TOC entry 3131 (class 0 OID 99023)
-- Dependencies: 222
-- Data for Name: users_bio; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".users_bio VALUES ('slimmy', 'get out');
INSERT INTO "issue#1".users_bio VALUES ('rembrandt', 'posh gorrila');
INSERT INTO "issue#1".users_bio VALUES ('loveless', 'i don&#39;t know what&#39;s real!');
INSERT INTO "issue#1".users_bio VALUES ('Isis Cane', 'War on your nation and fire to your idols, you scum!');


--
-- TOC entry 3214 (class 0 OID 0)
-- Dependencies: 225
-- Name: comments_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".comments_id_seq', 8, true);


--
-- TOC entry 3215 (class 0 OID 0)
-- Dependencies: 210
-- Name: feeds_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".feeds_id_seq', 12, true);


--
-- TOC entry 3216 (class 0 OID 0)
-- Dependencies: 215
-- Name: post_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".post_id_seq', 5, true);


--
-- TOC entry 3217 (class 0 OID 0)
-- Dependencies: 219
-- Name: title_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".title_id_seq', 55, true);


--
-- TOC entry 2900 (class 2606 OID 99033)
-- Name: channel_admins channel_admins_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_pkey PRIMARY KEY (channel_username, username);


--
-- TOC entry 2902 (class 2606 OID 99035)
-- Name: channel_official_catalog channel_catalog_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_pkey PRIMARY KEY (channel_username, release_id);


--
-- TOC entry 2938 (class 2606 OID 115137)
-- Name: channel_pictures channel_pictures_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_pictures
    ADD CONSTRAINT channel_pictures_pk PRIMARY KEY (channelname);


--
-- TOC entry 2904 (class 2606 OID 99037)
-- Name: channels channels_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channels
    ADD CONSTRAINT channels_pkey PRIMARY KEY (username);


--
-- TOC entry 2948 (class 2606 OID 123413)
-- Name: tsvs_comment comment_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_comment
    ADD CONSTRAINT comment_tsvs_pk PRIMARY KEY (comment_id);


--
-- TOC entry 2906 (class 2606 OID 123405)
-- Name: comments comments_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comments
    ADD CONSTRAINT comments_pk PRIMARY KEY (id);


--
-- TOC entry 2908 (class 2606 OID 99041)
-- Name: feed_subscriptions feed_channels_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feed_subscriptions
    ADD CONSTRAINT feed_channels_pkey PRIMARY KEY (feed_id, channel_username);


--
-- TOC entry 2910 (class 2606 OID 99043)
-- Name: feeds feeds_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feeds
    ADD CONSTRAINT feeds_pkey PRIMARY KEY (id);


--
-- TOC entry 2912 (class 2606 OID 99045)
-- Name: releases_image_based image based_image_name_key; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_image_based
    ADD CONSTRAINT "image based_image_name_key" UNIQUE (image_name);


--
-- TOC entry 2914 (class 2606 OID 99047)
-- Name: releases_image_based image based_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_image_based
    ADD CONSTRAINT "image based_pkey" PRIMARY KEY (release_id);


--
-- TOC entry 2916 (class 2606 OID 99049)
-- Name: release_metadata metadata_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".release_metadata
    ADD CONSTRAINT metadata_pkey PRIMARY KEY (release_id);


--
-- TOC entry 2918 (class 2606 OID 99051)
-- Name: post_contents post_contents_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_contents
    ADD CONSTRAINT post_contents_pkey PRIMARY KEY (post_id, release_id);


--
-- TOC entry 2920 (class 2606 OID 99053)
-- Name: posts post_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- TOC entry 2922 (class 2606 OID 99055)
-- Name: post_stars post_stars_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_stars
    ADD CONSTRAINT post_stars_pkey PRIMARY KEY (username, post_id);


--
-- TOC entry 2945 (class 2606 OID 123351)
-- Name: tsvs_posts posts_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_posts
    ADD CONSTRAINT posts_tsvs_pk PRIMARY KEY (post_id);


--
-- TOC entry 2924 (class 2606 OID 99059)
-- Name: releases release_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases
    ADD CONSTRAINT release_pkey PRIMARY KEY (id);


--
-- TOC entry 2941 (class 2606 OID 123374)
-- Name: tsvs_release release_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_release
    ADD CONSTRAINT release_tsvs_pk PRIMARY KEY (release_id);


--
-- TOC entry 2926 (class 2606 OID 99057)
-- Name: releases_text_based text based_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_text_based
    ADD CONSTRAINT "text based_pkey" PRIMARY KEY (release_id);


--
-- TOC entry 2936 (class 2606 OID 115124)
-- Name: user_avatars user_avatars_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_avatars
    ADD CONSTRAINT user_avatars_pk PRIMARY KEY (username);


--
-- TOC entry 2928 (class 2606 OID 99061)
-- Name: user_bookmarks user_bookmarks_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_bookmarks
    ADD CONSTRAINT user_bookmarks_pkey PRIMARY KEY (username, post_id);


--
-- TOC entry 2934 (class 2606 OID 99063)
-- Name: users_bio users_bio_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users_bio
    ADD CONSTRAINT users_bio_pk PRIMARY KEY (username);


--
-- TOC entry 2930 (class 2606 OID 99065)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 2932 (class 2606 OID 99067)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users
    ADD CONSTRAINT users_pkey PRIMARY KEY (username);


--
-- TOC entry 2946 (class 1259 OID 123419)
-- Name: comment_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX comment_ts_index ON "issue#1".tsvs_comment USING gin (vector);


--
-- TOC entry 2943 (class 1259 OID 123375)
-- Name: post_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX post_ts_index ON "issue#1".tsvs_posts USING gin (vector);


--
-- TOC entry 2939 (class 1259 OID 123343)
-- Name: release_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX release_ts_index ON "issue#1".tsvs_release USING gin (vector);


--
-- TOC entry 2942 (class 1259 OID 123326)
-- Name: release_tsvs_release_id_uindex; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE UNIQUE INDEX release_tsvs_release_id_uindex ON "issue#1".tsvs_release USING btree (release_id);


--
-- TOC entry 2979 (class 2620 OID 131613)
-- Name: comments comment_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER comment_insert_trigger AFTER INSERT ON "issue#1".comments FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_comment_trigger();


--
-- TOC entry 2978 (class 2620 OID 131612)
-- Name: comments comment_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER comment_update_trigger AFTER UPDATE ON "issue#1".comments FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_comment_trigger();


--
-- TOC entry 2981 (class 2620 OID 131617)
-- Name: release_metadata metadata_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER metadata_insert_trigger AFTER INSERT ON "issue#1".release_metadata FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_on_metadata_update_trigger();


--
-- TOC entry 2980 (class 2620 OID 131616)
-- Name: release_metadata metadata_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER metadata_update_trigger AFTER UPDATE ON "issue#1".release_metadata FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_on_metadata_update_trigger();


--
-- TOC entry 2983 (class 2620 OID 131615)
-- Name: posts post_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER post_insert_trigger AFTER INSERT ON "issue#1".posts FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_post_trigger();


--
-- TOC entry 2982 (class 2620 OID 131614)
-- Name: posts post_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER post_update_trigger AFTER UPDATE ON "issue#1".posts FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_post_trigger();


--
-- TOC entry 2985 (class 2620 OID 123310)
-- Name: users setup_user; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER setup_user AFTER INSERT ON "issue#1".users FOR EACH ROW EXECUTE FUNCTION "issue#1".setup_user();


--
-- TOC entry 2984 (class 2620 OID 131618)
-- Name: releases_text_based text_based_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER text_based_update_trigger AFTER UPDATE ON "issue#1".releases_text_based FOR EACH ROW EXECUTE FUNCTION "issue#1".tsv_text_based_update_trigger();


--
-- TOC entry 2949 (class 2606 OID 99068)
-- Name: channel_admins channel_admins_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_channel_username_fkey FOREIGN KEY (channel_username) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2950 (class 2606 OID 99073)
-- Name: channel_admins channel_admins_user_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_user_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2951 (class 2606 OID 99078)
-- Name: channel_official_catalog channel_catalog_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_channel_username_fkey FOREIGN KEY (channel_username) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2952 (class 2606 OID 99083)
-- Name: channel_official_catalog channel_catalog_post_from_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_post_from_id_fkey FOREIGN KEY (post_from_id) REFERENCES "issue#1".posts(id);


--
-- TOC entry 2953 (class 2606 OID 99088)
-- Name: channel_official_catalog channel_catalog_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_official_catalog
    ADD CONSTRAINT channel_catalog_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2974 (class 2606 OID 115138)
-- Name: channel_pictures channel_pictures_channels_username_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_pictures
    ADD CONSTRAINT channel_pictures_channels_username_fk FOREIGN KEY (channelname) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2977 (class 2606 OID 123414)
-- Name: tsvs_comment comment_tsvs_comments_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_comment
    ADD CONSTRAINT comment_tsvs_comments_id_fk FOREIGN KEY (comment_id) REFERENCES "issue#1".comments(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2956 (class 2606 OID 123313)
-- Name: comments comments_commented_by_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comments
    ADD CONSTRAINT comments_commented_by_fkey FOREIGN KEY (commented_by) REFERENCES "issue#1".users(username) ON UPDATE CASCADE NOT VALID;


--
-- TOC entry 2955 (class 2606 OID 99098)
-- Name: comments comments_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comments
    ADD CONSTRAINT comments_post_id_fkey FOREIGN KEY (post_from) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2957 (class 2606 OID 99103)
-- Name: feed_subscriptions feed_channels_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feed_subscriptions
    ADD CONSTRAINT feed_channels_channel_username_fkey FOREIGN KEY (channel_username) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2958 (class 2606 OID 99108)
-- Name: feed_subscriptions feed_channels_feed_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feed_subscriptions
    ADD CONSTRAINT feed_channels_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES "issue#1".feeds(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2959 (class 2606 OID 99113)
-- Name: feeds feeds_owner_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".feeds
    ADD CONSTRAINT feeds_owner_username_fkey FOREIGN KEY (owner_username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2961 (class 2606 OID 99123)
-- Name: release_metadata metadata_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".release_metadata
    ADD CONSTRAINT metadata_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON DELETE CASCADE;


--
-- TOC entry 2962 (class 2606 OID 99128)
-- Name: post_contents post_contents_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_contents
    ADD CONSTRAINT post_contents_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2963 (class 2606 OID 139805)
-- Name: post_contents post_contents_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_contents
    ADD CONSTRAINT post_contents_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2966 (class 2606 OID 99138)
-- Name: post_stars post_stars_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_stars
    ADD CONSTRAINT post_stars_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2967 (class 2606 OID 99143)
-- Name: post_stars post_stars_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".post_stars
    ADD CONSTRAINT post_stars_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2964 (class 2606 OID 99148)
-- Name: posts posts_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts
    ADD CONSTRAINT posts_channel_username_fkey FOREIGN KEY (channel_from) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2965 (class 2606 OID 99153)
-- Name: posts posts_poster_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts
    ADD CONSTRAINT posts_poster_username_fkey FOREIGN KEY (posted_by) REFERENCES "issue#1".users(username) ON UPDATE CASCADE NOT VALID;


--
-- TOC entry 2976 (class 2606 OID 123352)
-- Name: tsvs_posts posts_tsvs_posts_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_posts
    ADD CONSTRAINT posts_tsvs_posts_id_fk FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2975 (class 2606 OID 123332)
-- Name: tsvs_release release_tsvs_releases_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".tsvs_release
    ADD CONSTRAINT release_tsvs_releases_id_fk FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2960 (class 2606 OID 139810)
-- Name: releases_image_based releases_image_based_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_image_based
    ADD CONSTRAINT releases_image_based_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2968 (class 2606 OID 99158)
-- Name: releases releases_owner_channel_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases
    ADD CONSTRAINT releases_owner_channel_fkey FOREIGN KEY (owner_channel) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2954 (class 2606 OID 99163)
-- Name: channel_stickies stickied_posts_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_stickies
    ADD CONSTRAINT stickied_posts_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2969 (class 2606 OID 99168)
-- Name: releases_text_based text based_title_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases_text_based
    ADD CONSTRAINT "text based_title_id_fkey" FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON DELETE CASCADE NOT VALID;


--
-- TOC entry 2973 (class 2606 OID 115125)
-- Name: user_avatars user_avatars_users_username_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_avatars
    ADD CONSTRAINT user_avatars_users_username_fk FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2970 (class 2606 OID 99173)
-- Name: user_bookmarks user_bookmarks_post_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_bookmarks
    ADD CONSTRAINT user_bookmarks_post_id_fkey FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2971 (class 2606 OID 99178)
-- Name: user_bookmarks user_bookmarks_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".user_bookmarks
    ADD CONSTRAINT user_bookmarks_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 2972 (class 2606 OID 99183)
-- Name: users_bio users_bio_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".users_bio
    ADD CONSTRAINT users_bio_username_fkey FOREIGN KEY (username) REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 3146 (class 0 OID 0)
-- Dependencies: 263
-- Name: FUNCTION citextin(cstring); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextin(cstring) TO "issue#1_REST";


--
-- TOC entry 3147 (class 0 OID 0)
-- Dependencies: 229
-- Name: FUNCTION citextout("issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextout("issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3148 (class 0 OID 0)
-- Dependencies: 264
-- Name: FUNCTION citextrecv(internal); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextrecv(internal) TO "issue#1_REST";


--
-- TOC entry 3149 (class 0 OID 0)
-- Dependencies: 265
-- Name: FUNCTION citextsend("issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citextsend("issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3150 (class 0 OID 0)
-- Dependencies: 230
-- Name: FUNCTION citext(boolean); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext(boolean) TO "issue#1_REST";


--
-- TOC entry 3151 (class 0 OID 0)
-- Dependencies: 266
-- Name: FUNCTION citext(character); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext(character) TO "issue#1_REST";


--
-- TOC entry 3152 (class 0 OID 0)
-- Dependencies: 231
-- Name: FUNCTION citext(inet); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext(inet) TO "issue#1_REST";


--
-- TOC entry 3153 (class 0 OID 0)
-- Dependencies: 267
-- Name: FUNCTION citext_cmp("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_cmp("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3154 (class 0 OID 0)
-- Dependencies: 268
-- Name: FUNCTION citext_eq("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_eq("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3155 (class 0 OID 0)
-- Dependencies: 234
-- Name: FUNCTION citext_ge("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_ge("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3156 (class 0 OID 0)
-- Dependencies: 235
-- Name: FUNCTION citext_gt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_gt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3157 (class 0 OID 0)
-- Dependencies: 269
-- Name: FUNCTION citext_hash("issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_hash("issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3158 (class 0 OID 0)
-- Dependencies: 270
-- Name: FUNCTION citext_hash_extended("issue#1".citext, bigint); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_hash_extended("issue#1".citext, bigint) TO "issue#1_REST";


--
-- TOC entry 3159 (class 0 OID 0)
-- Dependencies: 236
-- Name: FUNCTION citext_larger("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_larger("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3160 (class 0 OID 0)
-- Dependencies: 271
-- Name: FUNCTION citext_le("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_le("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3161 (class 0 OID 0)
-- Dependencies: 232
-- Name: FUNCTION citext_lt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_lt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3162 (class 0 OID 0)
-- Dependencies: 233
-- Name: FUNCTION citext_ne("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_ne("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3163 (class 0 OID 0)
-- Dependencies: 272
-- Name: FUNCTION citext_pattern_cmp("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_cmp("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3164 (class 0 OID 0)
-- Dependencies: 273
-- Name: FUNCTION citext_pattern_ge("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_ge("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3165 (class 0 OID 0)
-- Dependencies: 274
-- Name: FUNCTION citext_pattern_gt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_gt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3166 (class 0 OID 0)
-- Dependencies: 275
-- Name: FUNCTION citext_pattern_le("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_le("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3167 (class 0 OID 0)
-- Dependencies: 249
-- Name: FUNCTION citext_pattern_lt("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_pattern_lt("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3168 (class 0 OID 0)
-- Dependencies: 237
-- Name: FUNCTION citext_smaller("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".citext_smaller("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3169 (class 0 OID 0)
-- Dependencies: 276
-- Name: FUNCTION regexp_match("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_match("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3170 (class 0 OID 0)
-- Dependencies: 277
-- Name: FUNCTION regexp_match("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_match("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3171 (class 0 OID 0)
-- Dependencies: 278
-- Name: FUNCTION regexp_matches("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_matches("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3172 (class 0 OID 0)
-- Dependencies: 243
-- Name: FUNCTION regexp_matches("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_matches("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3173 (class 0 OID 0)
-- Dependencies: 279
-- Name: FUNCTION regexp_replace("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_replace("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3174 (class 0 OID 0)
-- Dependencies: 280
-- Name: FUNCTION regexp_replace("issue#1".citext, "issue#1".citext, text, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_replace("issue#1".citext, "issue#1".citext, text, text) TO "issue#1_REST";


--
-- TOC entry 3175 (class 0 OID 0)
-- Dependencies: 244
-- Name: FUNCTION regexp_split_to_array("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_array("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3176 (class 0 OID 0)
-- Dependencies: 245
-- Name: FUNCTION regexp_split_to_array("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_array("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3177 (class 0 OID 0)
-- Dependencies: 246
-- Name: FUNCTION regexp_split_to_table("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_table("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3178 (class 0 OID 0)
-- Dependencies: 281
-- Name: FUNCTION regexp_split_to_table("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".regexp_split_to_table("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3179 (class 0 OID 0)
-- Dependencies: 247
-- Name: FUNCTION replace("issue#1".citext, "issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".replace("issue#1".citext, "issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3180 (class 0 OID 0)
-- Dependencies: 282
-- Name: FUNCTION split_part("issue#1".citext, "issue#1".citext, integer); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".split_part("issue#1".citext, "issue#1".citext, integer) TO "issue#1_REST";


--
-- TOC entry 3181 (class 0 OID 0)
-- Dependencies: 248
-- Name: FUNCTION strpos("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".strpos("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3182 (class 0 OID 0)
-- Dependencies: 238
-- Name: FUNCTION texticlike("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticlike("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3183 (class 0 OID 0)
-- Dependencies: 242
-- Name: FUNCTION texticlike("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticlike("issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3184 (class 0 OID 0)
-- Dependencies: 283
-- Name: FUNCTION texticnlike("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticnlike("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3185 (class 0 OID 0)
-- Dependencies: 239
-- Name: FUNCTION texticnlike("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticnlike("issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3186 (class 0 OID 0)
-- Dependencies: 284
-- Name: FUNCTION texticregexeq("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexeq("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3187 (class 0 OID 0)
-- Dependencies: 285
-- Name: FUNCTION texticregexeq("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexeq("issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3188 (class 0 OID 0)
-- Dependencies: 240
-- Name: FUNCTION texticregexne("issue#1".citext, "issue#1".citext); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexne("issue#1".citext, "issue#1".citext) TO "issue#1_REST";


--
-- TOC entry 3189 (class 0 OID 0)
-- Dependencies: 241
-- Name: FUNCTION texticregexne("issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".texticregexne("issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3190 (class 0 OID 0)
-- Dependencies: 250
-- Name: FUNCTION translate("issue#1".citext, "issue#1".citext, text); Type: ACL; Schema: issue#1; Owner: postgres
--

GRANT ALL ON FUNCTION "issue#1".translate("issue#1".citext, "issue#1".citext, text) TO "issue#1_REST";


--
-- TOC entry 3191 (class 0 OID 0)
-- Dependencies: 203
-- Name: TABLE channel_admins; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channel_admins TO "issue#1_REST";


--
-- TOC entry 3192 (class 0 OID 0)
-- Dependencies: 204
-- Name: TABLE channel_official_catalog; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channel_official_catalog TO "issue#1_REST";


--
-- TOC entry 3193 (class 0 OID 0)
-- Dependencies: 205
-- Name: TABLE channel_stickies; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channel_stickies TO "issue#1_REST";


--
-- TOC entry 3194 (class 0 OID 0)
-- Dependencies: 206
-- Name: TABLE channels; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".channels TO "issue#1_REST";


--
-- TOC entry 3195 (class 0 OID 0)
-- Dependencies: 207
-- Name: TABLE comments; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".comments TO "issue#1_REST";


--
-- TOC entry 3196 (class 0 OID 0)
-- Dependencies: 208
-- Name: TABLE feed_subscriptions; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".feed_subscriptions TO "issue#1_REST";


--
-- TOC entry 3197 (class 0 OID 0)
-- Dependencies: 209
-- Name: TABLE feeds; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".feeds TO "issue#1_REST";


--
-- TOC entry 3199 (class 0 OID 0)
-- Dependencies: 210
-- Name: SEQUENCE feeds_id_seq; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON SEQUENCE "issue#1".feeds_id_seq TO "issue#1_REST";


--
-- TOC entry 3200 (class 0 OID 0)
-- Dependencies: 213
-- Name: TABLE post_contents; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".post_contents TO "issue#1_REST";


--
-- TOC entry 3201 (class 0 OID 0)
-- Dependencies: 214
-- Name: TABLE posts; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".posts TO "issue#1_REST";


--
-- TOC entry 3203 (class 0 OID 0)
-- Dependencies: 215
-- Name: SEQUENCE post_id_seq; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON SEQUENCE "issue#1".post_id_seq TO "issue#1_REST";


--
-- TOC entry 3204 (class 0 OID 0)
-- Dependencies: 216
-- Name: TABLE post_stars; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".post_stars TO "issue#1_REST";


--
-- TOC entry 3205 (class 0 OID 0)
-- Dependencies: 212
-- Name: TABLE release_metadata; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".release_metadata TO "issue#1_REST";


--
-- TOC entry 3206 (class 0 OID 0)
-- Dependencies: 217
-- Name: TABLE releases; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".releases TO "issue#1_REST";


--
-- TOC entry 3207 (class 0 OID 0)
-- Dependencies: 211
-- Name: TABLE releases_image_based; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".releases_image_based TO "issue#1_REST";


--
-- TOC entry 3208 (class 0 OID 0)
-- Dependencies: 218
-- Name: TABLE releases_text_based; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".releases_text_based TO "issue#1_REST";


--
-- TOC entry 3210 (class 0 OID 0)
-- Dependencies: 219
-- Name: SEQUENCE title_id_seq; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON SEQUENCE "issue#1".title_id_seq TO "issue#1_REST";


--
-- TOC entry 3211 (class 0 OID 0)
-- Dependencies: 220
-- Name: TABLE user_bookmarks; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".user_bookmarks TO "issue#1_REST";


--
-- TOC entry 3212 (class 0 OID 0)
-- Dependencies: 221
-- Name: TABLE users; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".users TO "issue#1_REST";


--
-- TOC entry 3213 (class 0 OID 0)
-- Dependencies: 222
-- Name: TABLE users_bio; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".users_bio TO "issue#1_REST";


-- Completed on 2020-01-21 15:06:17

--
-- PostgreSQL database dump complete
--

