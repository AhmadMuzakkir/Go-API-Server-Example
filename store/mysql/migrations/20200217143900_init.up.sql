CREATE TABLE IF NOT EXISTS `users`
(
    `id`            INT          NOT NULL AUTO_INCREMENT,
    `username`      VARCHAR(255) NOT NULL,
    `password_hash` VARCHAR(255) NOT NULL,

    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_username` (`username`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `tokens`
(
    `user_id`    INT          NOT NULL,
    `token`      VARCHAR(255) NOT NULL,
    `updated_at` DATETIME(6)  NULL DEFAULT NULL,

    CONSTRAINT `fk_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    PRIMARY KEY (`token`),
    UNIQUE INDEX `idx_user_id` (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `messages`
(
    `id`         INT          NOT NULL AUTO_INCREMENT,
    `content`    VARCHAR(255) NOT NULL,
    `sender_id`  INT          NOT NULL,
    `created_at` DATETIME(6)  NULL DEFAULT NULL,
    `updated_at` DATETIME(6)  NULL DEFAULT NULL,

    CONSTRAINT `fk_messages_sender` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    PRIMARY KEY (`id`),
    INDEX `idx_messages_sender_id` (`sender_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `user_message_recipients`
(
    `message_id`   INT NOT NULL,
    `recipient_id` INT NOT NULL,

    CONSTRAINT `fk_user_message_recipients_message` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_message_recipients_user` FOREIGN KEY (`recipient_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    UNIQUE INDEX `idx_user_message_recipients` (`message_id`, `recipient_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;