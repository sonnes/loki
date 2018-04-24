CREATE DATABASE IF NOT EXISTS edgestore ENCODING 'UTF8';

CREATE TABLE IF NOT EXISTS follow (
  id varchar PRIMARY KEY,

  src_id bigint,
  src_type varchar,
  dest_id bigint,
  dest_type varchar,
  score decimal,
  data jsonb,
  status varchar,
  updated timestamp

);

CREATE INDEX follow_src_id ON follow (src_id);
CREATE INDEX follow_dest_id ON follow (dest_id);
CREATE INDEX follow_score ON follow (score);
CREATE INDEX follow_status ON follow (status);
CREATE INDEX follow_combi ON follow (src_id, dest_id, score, status);


CREATE TABLE IF NOT EXISTS follow_import (
  id varchar,

  src_id bigint,
  src_type varchar,
  dest_id bigint,
  dest_type varchar,
  score decimal,
  data jsonb,
  status varchar,
  updated timestamp

);
