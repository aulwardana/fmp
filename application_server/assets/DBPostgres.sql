﻿CREATE TABLE place
(
  uid serial NOT NULL,
  name character varying(255) NOT NULL,
  type character varying(255) NOT NULL,
  location character varying(500) NOT NULL,
  state character varying(255) NOT NULL,
  kawasan character varying(255) NOT NULL,
  latitude character varying(255) NOT NULL,
  longitude character varying(255) NOT NULL,
  code character varying(255) NOT NULL,
  CONSTRAINT place_pkey PRIMARY KEY (uid)
)
WITH (OIDS=FALSE);