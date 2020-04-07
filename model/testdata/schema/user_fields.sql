CREATE TABLE `user_fields` (
  `id` bigint(20) unsigned NOT NULL,
  `user_id` bigint(20) unsigned NOT NULL,
  `field_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_fields_01` (`user_id`, `field_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

