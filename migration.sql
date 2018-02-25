ALTER TABLE `container_items` DROP FOREIGN KEY `fk_container_items_containers1`;
ALTER TABLE `container_items` ADD FOREIGN KEY (`container_id`) REFERENCES `containers` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `locations` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `containers` DROP FOREIGN KEY `fk_containers_users`;
ALTER TABLE `containers` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE `api_users` DROP FOREIGN KEY `fk_user_id_constraint`;
ALTER TABLE `api_users` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
