
--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA "issue#1";


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: on_metadata_update_vcs_trigger(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".on_metadata_update_vcs_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    tsv tsvector := ''::tsvector;
BEGIN
    case (select r.type from releases as r where id = new.release_id)
        when 'image' then tsv := (SELECT setweight(to_tsvector('simple', COALESCE(owner_channel, '')), 'B')
                                             ||
                                         setweight(to_tsvector('simple', 'image'), 'A')
                                             ||
                                         setweight(to_tsvector('simple', COALESCE(new.description, '')), 'D')
                                             ||
                                         setweight(to_tsvector('simple', new.other), 'C')
                                             ||
                                         setweight(to_tsvector('simple', COALESCE(new.genre_defining, '')), 'B')
                                             ||
                                         setweight(to_tsvector('simple', COALESCE(new.title, '')), 'A')
                                  FROM (SELECT owner_channel
                                        FROM releases
                                        WHERE id = new.release_id) as r2
        );
                          insert into release_tsvs (release_id, vector)
                          values (new.release_id, tsv)
                          ON CONFLICT (release_id) do update
                              set vector= tsv;
        when 'text' then tsv := (SELECT setweight(to_tsvector('simple', COALESCE(owner_channel, '')), 'B')
                                            ||
                                        setweight(to_tsvector('simple', 'text'), 'A')
                                            ||
                                        setweight(to_tsvector('simple', COALESCE(new.description, '')), 'D')
                                            ||
                                        setweight(to_tsvector('simple', new.other), 'C')
                                            ||
                                        setweight(to_tsvector('simple', COALESCE(new.genre_defining, '')), 'B')
                                            ||
                                        setweight(to_tsvector('simple', COALESCE(new.title, '')), 'A')
                                            ||
                                        setweight(to_tsvector('simple', COALESCE(content, '')), 'D')
                                 FROM (SELECT id as release_id, owner_channel
                                       FROM releases
                                       WHERE id = new.release_id
                                      ) as r2
                                          left join text_based on text_based.release_id = r2.release_id
        );
                         insert into release_tsvs (release_id, vector)
                         values (new.release_id, tsv)
                         ON CONFLICT (release_id) do update
                             set vector= tsv;
        else
        end case;
    return new;
END;
$$;


ALTER FUNCTION "issue#1".on_metadata_update_vcs_trigger() OWNER TO "issue#1_dev";

--
-- Name: post_tcs_trigger(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".post_tcs_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    tsv tsvector := ''::tsvector;
BEGIN
    tsv := (SELECT setweight(to_tsvector('simple', COALESCE(new.description, '')), 'D')
                       ||
                   setweight(to_tsvector('simple', COALESCE(new.title, '')), 'A')
                       ||
                   setweight(to_tsvector('simple', COALESCE(new.posted_by, '')), 'B')
                       ||
                   setweight(to_tsvector('simple', COALESCE(new.channel_from, '')), 'B')
    );
    insert into posts_tsvs (post_id, vector)
    values (new.id, tsv)
    ON CONFLICT (post_id)
        do update
        set vector= tsv;
    return new;
END;
$$;


ALTER FUNCTION "issue#1".post_tcs_trigger() OWNER TO "issue#1_dev";

--
-- Name: setup_user(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".setup_user() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    insert into channels(username, name)
    values (new.username, new.username || '''s channel');
    insert into feeds(owner_username, sorting)
    values  (new.username, 'hot');
    return new;
END;
$$;


ALTER FUNCTION "issue#1".setup_user() OWNER TO "issue#1_dev";

--
-- Name: text_based_update_vcs_trigger(); Type: FUNCTION; Schema: issue#1; Owner: issue#1_dev
--

CREATE FUNCTION "issue#1".text_based_update_vcs_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    tsv tsvector := ''::tsvector;
BEGIN
    tsv := (SELECT setweight(to_tsvector('simple', COALESCE(owner_channel, '')), 'B')
                       ||
                   setweight(to_tsvector('simple', 'text'), 'A')
                       ||
                   setweight(to_tsvector('simple', COALESCE(description, '')), 'D')
                       ||
                   setweight(to_tsvector('simple', other), 'C')
                       ||
                   setweight(to_tsvector('simple', COALESCE(genre_defining, '')), 'B')
                       ||
                   setweight(to_tsvector('simple', COALESCE(title, '')), 'A')
                       ||
                   setweight(to_tsvector('simple', COALESCE(new.content, '')), 'D')
            FROM (SELECT id as release_id, owner_channel
                  FROM releases
                  WHERE id = new.release_id
                 ) as r2
                     left join metadata on metadata.release_id = r2.release_id
    );
    insert into release_tsvs (release_id, vector)
    values (new.release_id, tsv)
    ON CONFLICT (release_id) do update
        set vector= tsv;
    return new;
END;
$$;


ALTER FUNCTION "issue#1".text_based_update_vcs_trigger() OWNER TO "issue#1_dev";

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: channel_admins; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".channel_admins (
    channel_username character varying(24) NOT NULL,
    "user" character varying(24) NOT NULL,
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
-- Name: comment_tsvs; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".comment_tsvs (
    comment_id integer NOT NULL,
    vector tsvector
);


ALTER TABLE "issue#1".comment_tsvs OWNER TO "issue#1_dev";

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
-- Name: image_based; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".image_based (
    release_id integer NOT NULL,
    image_name text NOT NULL
);


ALTER TABLE "issue#1".image_based OWNER TO "issue#1_dev";

--
-- Name: metadata; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".metadata (
    release_id integer NOT NULL,
    description text,
    other jsonb,
    genre_defining text,
    release_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    title text
);


ALTER TABLE "issue#1".metadata OWNER TO "issue#1_dev";

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
-- Name: posts_tsvs; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".posts_tsvs (
    post_id integer NOT NULL,
    vector tsvector
);


ALTER TABLE "issue#1".posts_tsvs OWNER TO "issue#1_dev";

--
-- Name: release_tsvs; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".release_tsvs (
    release_id integer NOT NULL,
    vector tsvector
);


ALTER TABLE "issue#1".release_tsvs OWNER TO "issue#1_dev";

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
-- Name: text_based; Type: TABLE; Schema: issue#1; Owner: issue#1_dev
--

CREATE TABLE "issue#1".text_based (
    release_id integer NOT NULL,
    content text NOT NULL
);


ALTER TABLE "issue#1".text_based OWNER TO "issue#1_dev";

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



--
-- Data for Name: channel_official_catalog; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".channel_official_catalog VALUES ('chromagnum', 6, 4);


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


--
-- Data for Name: comment_tsvs; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- Data for Name: comments; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 1, -1, 'Epistien didn''t kill himself!', 'slimmy', '2020-01-11 19:29:29.058057+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 2, 1, 'duhh', 'rembrandtian', '2020-01-11 19:30:04.273774+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 3, -1, '<p>they can hear me crying…</p>
', 'loveless', '2020-01-11 20:16:33.519527+03');
INSERT INTO "issue#1".comments OVERRIDING SYSTEM VALUE VALUES (5, 6, 3, '<p>the diamond mines are <em>burning</em></p>
', 'rembrandtian', '2020-01-11 20:26:34.222721+03');


--
-- Data for Name: feed_subscriptions; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".feed_subscriptions VALUES (3, 'chromagnum', '2020-01-06 17:01:08.862048+03');


--
-- Data for Name: feeds; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".feeds VALUES ('rembrandt', 'hot', 3);
INSERT INTO "issue#1".feeds VALUES ('rembrandtian', 'top', 6);
INSERT INTO "issue#1".feeds VALUES ('slimmy', 'top', 7);
INSERT INTO "issue#1".feeds VALUES ('loveless', 'new', 5);


--
-- Data for Name: image_based; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".image_based VALUES (53, 'release.1eb2969c-cf4c-40cb-9b78-39bea2b089da.the image.jpg');


--
-- Data for Name: metadata; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".metadata VALUES (6, 'tips baby, tips', '{"genres": ["Academic"], "authors": ["@rembrandt"]}', 'Guide', '2020-01-04 11:25:15.055178+03', 'Tips1.md');
INSERT INTO "issue#1".metadata VALUES (52, 'Forget this not.', '{"genres": ["Omnious Message", "Prophecy"], "authors": ["Hooded Messenger"]}', 'Message', '2020-01-05 13:14:16.017584+03', 'The Journey Ends!');
INSERT INTO "issue#1".metadata VALUES (53, 'a broken moon', '{"genres": ["Cartoon"], "authors": ["Rebecca Sugar"]}', 'Cartoon', '2016-01-04 23:06:16.017584+03', 'Please don''t leave Pink, don''t leave please.');
INSERT INTO "issue#1".metadata VALUES (54, 'Guidelines to the future and to the deep honest, archaic path.', '{"genres": ["K-Pop"], "authors": ["Hooded Messenger"]}', 'Literature', '2020-12-05 13:14:16.017584+03', 'Above & Not There Yet');


--
-- Data for Name: post_contents; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- Data for Name: post_stars; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--



--
-- Data for Name: posts; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".posts VALUES (4, 'tipstitipst[tipsdtst]tipstitstipstsi[ps]', 'HEREWEGOLANG', 'rembrandt', 'chromagnum', '2020-01-04 20:58:27.926859+03');
INSERT INTO "issue#1".posts VALUES (3, 'B7 Chord, G6 Chord.', 'its so strangeeee', 'rembrandt', 'chromagnum', '2019-12-29 19:59:39.568527+03');
INSERT INTO "issue#1".posts VALUES (6, 'Welcome!', 'Issue #1 v0.1', 'slimmy', 'chromagnum', '2020-01-12 18:54:48.460537+03');
INSERT INTO "issue#1".posts VALUES (5, NULL, 'cordova-dovadovadovedova', 'rembrandt', 'chromagnum', '2019-12-29 20:00:53.714587+03');


--
-- Data for Name: posts_tsvs; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".posts_tsvs VALUES (4, '''chromagnum'':7B ''herewegolang'':5A ''ps'':4 ''rembrandt'':6B ''tipsdtst'':2 ''tipstitipst'':1 ''tipstitstipstsi'':3');
INSERT INTO "issue#1".posts_tsvs VALUES (3, '''b7'':1 ''chord'':2,4 ''chromagnum'':9B ''g6'':3 ''its'':5A ''rembrandt'':8B ''so'':6A ''strangeeee'':7A');
INSERT INTO "issue#1".posts_tsvs VALUES (6, '''1'':3A ''chromagnum'':6B ''issue'':2A ''slimmy'':5B ''v0.1'':4A ''welcome'':1');
INSERT INTO "issue#1".posts_tsvs VALUES (5, '''chromagnum'':5B ''cordova'':2A ''cordova-dovadovadovedova'':1A ''dovadovadovedova'':3A ''rembrandt'':4B');


--
-- Data for Name: release_tsvs; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".release_tsvs VALUES (52, '''27'':25 ''at'':24 ''beckons'':19 ''call'':23 ''chromagnum'':1B ''clock'':27 ''edge'':18 ''ends'':16A ''forget'':3 ''heed'':20 ''hooded'':11C ''it'':21 ''journey'':15A ''message'':7C,13B ''messenger'':12C ''not'':5 ''o'':26 ''omnious'':6C ''prophecy'':9C ''s'':22 ''text'':2A ''the'':14A,17 ''this'':4');
INSERT INTO "issue#1".release_tsvs VALUES (53, '''a'':3 ''broken'':4 ''cartoon'':6C,10B ''chromagnum'':1B ''don'':12A,16A ''image'':2A ''leave'':14A,18A ''moon'':5 ''pink'':15A ''please'':11A,19A ''rebecca'':8C ''sugar'':9C ''t'':13A,17A');
INSERT INTO "issue#1".release_tsvs VALUES (6, '''/go/src/github.com'':20 ''/go/src/github.com/slim-crown/issue-1-rest'':66,99 ''/gorilla/mux'':131 ''/slim-crown/issue-1-rest.git'':29 ''/slim-crown/issue-1-website.git'':34 ''/users'':293,302,311,321,336,947,951,954,957,961,965 ''0'':346 ''1.sql'':187 ''10'':1053 ''11'':991 ''12'':999 ''13'':1007 ''14'':1020 ''20'':1045 ''25'':344 ''a'':90,214,238,253,362,435,442,449,525,545,692,758,792,824,1056 ''able'':812 ''about'':418,645 ''academic'':6C ''according'':986 ''add'':995 ''adduser'':554 ''administrator'':140 ''afraid'':785 ''after'':248 ''ahead'':224,247 ''all'':203,327,459,637,706 ''along'':658 ''also'':236,330,691,738,1040 ''and'':217,235,244,260,323,353,380,400,415,473,533,625,648,685,702,803,806,819,985,1022,1026 ''any'':233,419,475,510 ''api'':979 ''application'':1013 ''applications'':994,1002,1005 ''are'':37,389,671,744 ''area'':174 ''article'':209,228 ''as'':139,354,668,741,976,1059 ''asc'':342 ''ask'':48,229,416,506,1036 ''at'':46,306,315,410 ''attempt'':149 ''authenticate'':854,898,967 ''authorize'':920,940 ''b'':69 ''baby'':4 ''bar'':164,182 ''based'':750 ''bash'':18 ''basic'':254,604 ''be'':159,406,434,663,725,779,811,840 ''before'':478,503,688 ''beleive'':491 ''besides'':602,682 ''between'':762 ''big'':1057 ''bookmarkpost'':589,600 ''bool'':587 ''boom'':827 ''branch'':60,64,77,91,97,108,122,216 ''branching'':208 ''but'':796,866 ''by'':520,1016,1069 ''cache'':901 ''caching'':760 ''called'':79 ''calls'':673,746 ''can'':158 ''cd'':19,65,98 ''checkout'':68,101 ''chromagnum'':1B ''clone'':26,31 ''cloning'':15 ''commit'':439,505,696,768 ''commiting'':448 ''compared'':1046 ''compile'':454 ''completely'':836 ''concurrency'':1008,1018 ''confident'':263 ''constructs'':420 ''copy'':167 ''create'':57,213,294,823 ''creation'':339 ''creation-time'':338 ''crown'':24 ''crud'':605 ''db'':749 ''deadline'':1072 ''define'':458 ''defining'':649 ''delete'':320,892,914,935,960 ''deleteuser'':569 ''deliver'':1066 ''dependencies'':126 ''design'':243 ''did'':414,595,632,800 ''different'':841 ''direct'':745 ''discuss'':626 ''do'':183,218,447,480,517,680,821 ''does'':154 ''doesn'':452 ''don'':197,221,422,678 ''done'':88,116,428,485,830,837,971 ''e.g'':291,852 ''each'':845 ''earler'':731 ''easier'':494 ''easiest'':665 ''easy'':377,740 ''end'':789 ''endpoints'':272 ''enough'':39,264 ''entites'':608 ''entities'':550 ''error'':403,557,562,568,572,582,588,594 ''errors'':477,795 ''especially'':324 ''essential'':190 ''evaluation'':1042,1049 ''expect'':288 ''fails'':143 ''fast'':975 ''feature'':71,81,103 ''feature-user-service'':70,80,102 ''features'':997,1068 ''feed'':120 ''feed-sercvice'':119 ''feel'':262 ''fi'':45 ''file'':241 ''first'':109,148,201,352,808,1071 ''fix'':474 ''following'':873 ''for'':185,273,469,851,855,869,1009 ''form'':300 ''found'':92,160 ''from'':61,117,169,269,296,317,730 ''function'':619 ''functions'':412,944 ''get'':128,301,304,325,335,883,886,905,908,926,929,946,950,970 ''getting'':484 ''getuser'':558 ''git'':14,17,25,30,67,100,110,207 ''github'':94 ''github.com'':28,33,130 ''github.com/gorilla/mux'':129 ''github.com/slim-crown/issue-1-rest.git'':27 ''github.com/slim-crown/issue-1-website.git'':32 ''given'':297,318 ''globals.sql'':170 ''go'':127,223,356,378,715 ''going'':278 ''good'':436,693 ''guide'':9B ''guideline'':11 ''half'':982 ''handle'':850 ''handled'':290 ''handler'':943 ''handlers'':271,351,388,431,468,488,635,802 ''handling'':404 ''hard'':1063 ''hardest'':781 ''have'':42,51,232,252,361,372,509,524,848,862 ''he'':1034 ''help'':54 ''home'':47 ''how'':55,84,1030 ''http'':942 ''i'':50,210,413,481,490,783,1043 ''if'':35,230,507 ''ii'':1050 ''implememt'':1039 ''implement'':280,349,396,532,636,667,769 ''implementation'':719,751 ''implementing'':726 ''improve'':1011 ''improvement'':1014 ''in'':308,734,761 ''includes'':331 ''input'':683 ''int'':580,593 ''inteface'':387,729 ''interface'':472,536,553,657,756 ''internationalization'':1021,1024 ''into'':123 ''is'':78,189,690,737,980,1044,1052 ''issue'':186 ''it'':141,153,432,513 ''its'':441,492 ''john'':348 ''json'':298,309,319 ''just'':281,497 ''know'':1029 ''later'':870 ''lighting'':177 ''like'':541,630,687,853 ''likely'':611 ''likley'':778 ''limit'':343,578 ''list'':282,363,526 ''little'':759 ''ll'':359,724,810,861 ''localization'':1025,1027 ''logic'':399 ''look'':409,539 ''lot'':793 ''lots'':401 ''lucky'':798 ''m'':784 ''make'':375,455,498,697,763 ''mates'':628 ''may'':538 ''me'':49 ''memory'':718 ''memory/cache'':735 ''menu'':163 ''merge'':86,111 ''merged'':114 ''method'':546 ''methods'':365,383,462,528,601,606,614,639,650,670,709,743,816,843,881,903,924 ''midway'':200 ''might'':788,1035 ''mkdir'':21 ''most'':610,669,742,777 ''much'':493,681,1031 ''need'':136,368,617,653 ''new'':76,215 ''next'':63,713 ''not'':40,835 ''note'':831 ''notice'':597 ''now'':489,512,521,629,689 ''of'':202,256,364,391,402,527,623,720,752,794,1032 ''ofcourse'':832 ''offset'':345,579 ''on'':89,93,146,161,179,220,326,379,384,466,640,654,674,714,865 ''one'':357,390 ''only'':981 ''open'':16,106,133,145,155 ''option'':334 ''or'':299 ''order'':188 ''origin/feature-feed-service'':112 ''origin/next'':74 ''other'':476,968 ''our'':332,1067 ''ours'':1060 ''out'':283 ''own'':59,125 ''part'':394 ''passhash'':585 ''passhashiscorrect'':583,599 ''pattern'':347,574 ''performance'':1010,1015 ''pgadmin'':134 ''plan'':245,876 ''post'':292,953,964 ''postgre'':922 ''postgresql'':774 ''postid'':592 ''pressed'':1064 ''project'':12,191,258,450,501,700,766,1041,1048,1058 ''pull'':825 ''put'':310,956 ''query'':156,172,176 ''questions'':234,417,511 ''re'':277,427,829,834,1062 ''read'':193,204 ''reading'':226 ''really'':443,482 ''reccommend'':483 ''rembrandt'':8C ''repository'':647,656,676,708,722,728,755,772,902,923 ''request'':826 ''require'':612 ''rest'':978 ''right'':817 ''run'':138,175 ''runs'':502,701,767 ''s'':286,514,868 ''same'':184,754,771 ''search'':333,895,917,937 ''searchuser'':573 ''securing'':992 ''security'':996 ''sent'':211 ''sercvice'':121,805 ''service'':73,83,105,275,370,386,461,471,519,535,543,548,552,642,846,858,880 ''services'':661 ''setup'':13,874 ''shit'':974 ''should'':374,523,531,537,662 ''simple'':239,672,733 ''since'':440,797 ''slim'':23 ''slim-crown'':22 ''so'':150,405 ''some'':624 ''something'':540 ''soon'':360 ''sortby'':337,575 ''sortorder'':341,576 ''special'':613 ''specialized'':842 ''specifed'':705 ''specifies'':544 ''specify'':381 ''sql'':132 ''start'':266,268 ''step'':782 ''stop'':199 ''string'':560,565,571,577,586,591 ''structure'':259 ''stuff'':686,807,969 ''succesful'':206 ''such'':1055 ''sucky'':444 ''sure'':407,456,499,698,764 ''syllabus'':989 ''symbol'':178 ''t'':198,222,453,679 ''test'':814,1003 ''testing'':1000 ''text'':2A,168,173,240 ''that'':53,227,366,451,463,495,529,615,651,677,710,822,844,859,867 ''the'':62,75,118,147,162,180,194,205,257,270,274,284,350,382,392,411,430,460,467,487,500,518,603,634,638,641,646,655,659,664,675,699,707,717,721,727,748,753,765,770,780,801,804,856,872,977,983,988,1070 ''then'':820 ''ther'':598 ''there'':818,838 ''these'':1033 ''they'':616 ''thing'':196,250,445 ''things'':246,376 ''think'':622 ''this'':542,736,775,973 ''time'':340,437,515,694 ''tips'':3,5,192 ''tips1.md'':10A ''to'':41,56,85,95,137,144,171,212,242,265,279,289,371,395,408,438,446,457,516,547,618,621,666,695,716,747,813,849,863,875,987,1038,1047,1065 ''tool'':157,181 ''tools'':52,166 ''toughest'':393 ''tricky'':398 ''try'':151,620 ''type'':551 ''u'':566 ''under'':165 ''understand'':424 ''understanding'':255 ''unlucky'':38 ''until'':152 ''up'':790 ''update'':889,911,932 ''update/create'':313 ''updateuser'':563 ''url'':285 ''us'':1037 ''use'':237,871,1023 ''used'':465,712 ''user'':72,82,104,295,305,314,549,555,556,561,567,581,857,884,890,893,896,899,906,912,915,918,921,927,933,936,938,941 ''username'':303,307,312,316,322,559,564,570,584,590,948,958,962,966 ''users'':328,887,909,930 ''using'':732,773,1017 ''usually'':142 ''validation'':684 ''ve'':704 ''very'':739 ''way'':496,660 ''we'':1061 ''web'':993,1001,1004,1012 ''week'':990,998,1006,1019 ''when'':425 ''where'':786 ''which'':329,373,1051 ''who'':1028 ''whole'':195 ''wi'':44 ''wi-fi'':43 ''will'':609,776,839,847 ''with'':397,429,486,627,633,757,791,972,1054 ''without'':225,643 ''work'':87,115,219,267,864,878,984 ''worrying'':644 ''would'':433 ''x'':879,882,885,888,891,894,897,900,904,907,910,913,916,919,925,928,931,934,939,945,949,952,955,959,963 ''y'':423 ''you'':36,113,135,231,249,251,261,276,287,355,358,367,421,426,464,479,504,508,522,530,596,631,652,703,711,723,787,799,809,828,833,860 ''your'':58,96,107,124,369,385,470,534,607,815,877');
INSERT INTO "issue#1".release_tsvs VALUES (54, '''above'':21A ''and'':7 ''archaic'':12 ''blue'':36 ''chromagnum'':1B ''deep'':10 ''do'':32 ''feathers'':29 ''future'':6 ''guidelines'':3 ''honest'':11 ''hooded'':18C ''i'':31 ''k'':15C ''k-pop'':14C ''literature'':20B ''messenger'':19C ''my'':28 ''no'':35 ''not'':22A,33 ''path'':13 ''pop'':16C ''ruffles'':27 ''see'':34 ''sky'':37 ''text'':2A ''the'':5,9,25 ''there'':23A ''to'':4,8 ''wind'':26 ''yet'':24A,30');


--
-- Data for Name: releases; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".releases VALUES (6, 'chromagnum', 'text', '2020-01-04 11:23:04.271553+03');
INSERT INTO "issue#1".releases VALUES (52, 'chromagnum', 'text', '2020-01-05 13:18:43.7471+03');
INSERT INTO "issue#1".releases VALUES (53, 'chromagnum', 'image', '2020-01-05 13:55:05.730832+03');
INSERT INTO "issue#1".releases VALUES (54, 'chromagnum', 'text', '2020-01-12 18:05:22.383129+03');


--
-- Data for Name: text_based; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".text_based VALUES (52, 'THE EDGE BECKONS! Heed it''s call at 27 o''clock.');
INSERT INTO "issue#1".text_based VALUES (6, '
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

### sql
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
This will most likley be the hardest step I''m afraid where you might end up with a lot of errors. But, since lucky you did the handlers and the sercvice and stuff first, you''ll be able to test your methods right there and then. 

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
INSERT INTO "issue#1".text_based VALUES (54, 'The wind ruffles my feathers. Yet I do not see no blue sky...');


--
-- Data for Name: user_avatars; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".user_avatars VALUES ('loveless', 'user.66cbf2be-c0ef-4056-8d43-aa7dca3e718b.lovelessness.jpg');


--
-- Data for Name: user_bookmarks; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".user_bookmarks VALUES ('rembrandtian', 5, '2019-12-29 20:28:43.700899+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('rembrandtian', 3, '2019-12-29 20:29:07.679261+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 3, '2019-12-29 20:35:52.794761+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 4, '2020-01-12 10:12:37.060716+03');
INSERT INTO "issue#1".user_bookmarks VALUES ('slimmy', 5, '2020-01-12 10:12:43.839661+03');


--
-- Data for Name: users; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".users VALUES ('yoftie3@gmail.com', 'rembrandt', '2019-12-28 22:52:07.479752+03', 'e20d91786b7a9317bb0c84c11af36f3f459adb1c386347aab05b68e9eac8dd287689ef1924dd80309c67d8656a9bf520cc4c8dea7fe0196d675437d21eb1da20', 'Y.', 'A.', 'Knowe');
INSERT INTO "issue#1".users VALUES ('hot@hotter.hottest', 'rembrandtian', '2019-12-28 22:54:02.72085+03', 'afe6d7d32ac4fff5eba8f09debdaaf57f946bd12239409f0f4b161558c7988f17a7cc47bc099bf9afa0f951458609055acd0149855f0ac7b604ecbeb28a63f5d', 'death', NULL, NULL);
INSERT INTO "issue#1".users VALUES ('yat@yayat.yat', 'slimmy', '2019-12-28 23:05:11.662742+03', '58df5e6ebc6c6a8b4760f514652d7baaa864cc1e7e656a5734a253ad3b16c4a69561a202f3ae79bab9cbadc6ed20224b90ad7d467a3bfb773964a80e27d03e97', 'Kebab', 'Bab', 'Bob');
INSERT INTO "issue#1".users VALUES ('stars@destination.com', 'loveless', '2020-01-08 01:49:31.153913+03', '87bba9f7d30eb9f64bc2d77eace67147c370e7073d22cb5bedfc5ae8b5bd3ca60b1748e21df9626749c031208e9811e38c3204a7efdf8a867ebfb9b066325251', 'Jeff', 'k.', 'Shoes');


--
-- Data for Name: users_bio; Type: TABLE DATA; Schema: issue#1; Owner: issue#1_dev
--

INSERT INTO "issue#1".users_bio VALUES ('slimmy', 'get out');
INSERT INTO "issue#1".users_bio VALUES ('rembrandt', 'posh gorrila');
INSERT INTO "issue#1".users_bio VALUES ('loveless', 'i don&#39;t know what&#39;s real!');


--
-- Name: comments_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".comments_id_seq', 7, true);


--
-- Name: feeds_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".feeds_id_seq', 7, true);


--
-- Name: post_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".post_id_seq', 5, true);


--
-- Name: title_id_seq; Type: SEQUENCE SET; Schema: issue#1; Owner: issue#1_dev
--

SELECT pg_catalog.setval('"issue#1".title_id_seq', 55, true);


--
-- Name: channel_admins channel_admins_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_pkey PRIMARY KEY (channel_username, "user");


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
-- Name: comment_tsvs comment_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comment_tsvs
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
-- Name: image_based image based_image_name_key; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".image_based
    ADD CONSTRAINT "image based_image_name_key" UNIQUE (image_name);


--
-- Name: image_based image based_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".image_based
    ADD CONSTRAINT "image based_pkey" PRIMARY KEY (release_id);


--
-- Name: metadata metadata_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".metadata
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
-- Name: posts_tsvs posts_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts_tsvs
    ADD CONSTRAINT posts_tsvs_pk PRIMARY KEY (post_id);


--
-- Name: releases release_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".releases
    ADD CONSTRAINT release_pkey PRIMARY KEY (id);


--
-- Name: release_tsvs release_tsvs_pk; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".release_tsvs
    ADD CONSTRAINT release_tsvs_pk PRIMARY KEY (release_id);


--
-- Name: text_based text based_pkey; Type: CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".text_based
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

CREATE INDEX comment_ts_index ON "issue#1".comment_tsvs USING gin (vector);


--
-- Name: post_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX post_ts_index ON "issue#1".posts_tsvs USING gin (vector);


--
-- Name: release_ts_index; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE INDEX release_ts_index ON "issue#1".release_tsvs USING gin (vector);


--
-- Name: release_tsvs_release_id_uindex; Type: INDEX; Schema: issue#1; Owner: issue#1_dev
--

CREATE UNIQUE INDEX release_tsvs_release_id_uindex ON "issue#1".release_tsvs USING btree (release_id);


--
-- Name: metadata metadata_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER metadata_insert_trigger AFTER INSERT ON "issue#1".metadata FOR EACH ROW EXECUTE FUNCTION "issue#1".on_metadata_update_vcs_trigger();


--
-- Name: metadata metadata_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER metadata_update_trigger AFTER UPDATE ON "issue#1".metadata FOR EACH ROW EXECUTE FUNCTION "issue#1".on_metadata_update_vcs_trigger();


--
-- Name: posts post_insert_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER post_insert_trigger AFTER INSERT ON "issue#1".posts FOR EACH ROW EXECUTE FUNCTION "issue#1".post_tcs_trigger();


--
-- Name: posts post_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER post_update_trigger AFTER UPDATE ON "issue#1".posts FOR EACH ROW EXECUTE FUNCTION "issue#1".post_tcs_trigger();


--
-- Name: users setup_user; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER setup_user AFTER INSERT ON "issue#1".users FOR EACH ROW EXECUTE FUNCTION "issue#1".setup_user();


--
-- Name: text_based text_based_update_trigger; Type: TRIGGER; Schema: issue#1; Owner: issue#1_dev
--

CREATE TRIGGER text_based_update_trigger AFTER UPDATE ON "issue#1".text_based FOR EACH ROW EXECUTE FUNCTION "issue#1".text_based_update_vcs_trigger();


--
-- Name: channel_admins channel_admins_channel_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_channel_username_fkey FOREIGN KEY (channel_username) REFERENCES "issue#1".channels(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- Name: channel_admins channel_admins_user_username_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".channel_admins
    ADD CONSTRAINT channel_admins_user_username_fkey FOREIGN KEY ("user") REFERENCES "issue#1".users(username) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


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
-- Name: comment_tsvs comment_tsvs_comments_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".comment_tsvs
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
-- Name: image_based fk to title id; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".image_based
    ADD CONSTRAINT "fk to title id" FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON DELETE CASCADE NOT VALID;


--
-- Name: metadata metadata_release_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".metadata
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
    ADD CONSTRAINT post_contents_release_id_fkey FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE;


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
-- Name: posts_tsvs posts_tsvs_posts_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".posts_tsvs
    ADD CONSTRAINT posts_tsvs_posts_id_fk FOREIGN KEY (post_id) REFERENCES "issue#1".posts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: release_tsvs release_tsvs_releases_id_fk; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".release_tsvs
    ADD CONSTRAINT release_tsvs_releases_id_fk FOREIGN KEY (release_id) REFERENCES "issue#1".releases(id) ON UPDATE CASCADE ON DELETE CASCADE;


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
-- Name: text_based text based_title_id_fkey; Type: FK CONSTRAINT; Schema: issue#1; Owner: issue#1_dev
--

ALTER TABLE ONLY "issue#1".text_based
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
-- Name: TABLE image_based; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".image_based TO "issue#1_REST";


--
-- Name: TABLE metadata; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".metadata TO "issue#1_REST";


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
-- Name: TABLE releases; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".releases TO "issue#1_REST";


--
-- Name: TABLE text_based; Type: ACL; Schema: issue#1; Owner: issue#1_dev
--

GRANT ALL ON TABLE "issue#1".text_based TO "issue#1_REST";


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

