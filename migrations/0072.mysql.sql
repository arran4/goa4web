CREATE TABLE `role_subscription_archetypes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `role_id` int NOT NULL,
  `archetype_name` varchar(128) NOT NULL,
  `pattern` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `role_subscription_archetypes_role_idx` (`role_id`)
);
