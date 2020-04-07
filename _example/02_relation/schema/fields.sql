CREATE TABLE `fields` (
  `id` bigint(20) unsigned NOT NULL,
  `name` varchar(30) DEFAULT NULL,
  `location_x` int NOT NULL,
  `location_y` int NOT NULL,
  `object_num` int NOT NULL,
  `level` int NOT NULL,
  `difficulty` int NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_fields_01` (`name`),
  UNIQUE KEY `uq_fields_02` (`location_x`, `location_y`),
  KEY `idx_fields_03` (`object_num`),
  KEY `idx_fields_04` (`difficulty`, `level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

