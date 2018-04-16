CREATE DATABASE edgestore ENCODING 'UTF8';

CREATE TABLE IF NOT EXISTS test_edge (
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

CREATE INDEX test_edge_src_id ON feed (src_id);
CREATE INDEX test_edge_dest_id ON feed (dest_id);
CREATE INDEX test_edge_score ON feed (score);
CREATE INDEX test_edge_status ON feed (status);
CREATE INDEX test_edge_combi ON feed (src_id, dest_id, score, status);


