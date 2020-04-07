CREATE TABLE `users` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(30) DEFAULT NULL,
  `sex` enum('man','woman') NOT NULL,
  `age` int NOT NULL,
  `skill_id` bigint(20) unsigned NOT NULL,
  `skill_rank` int NOT NULL,
  `group_id` bigint(20) unsigned NOT NULL,
  `world_id` bigint(20) unsigned NOT NULL,
  `field_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_users_01` (`name`),
  UNIQUE KEY `uq_users_02` (`skill_id`, `skill_rank`),
  KEY `idx_users_03` (`group_id`),
  KEY `idx_users_04` (`world_id`, `field_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

