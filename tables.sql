CREATE DATABASE `budgli` /*!40100 DEFAULT CHARACTER SET utf8mb4 */;

-- budgli.current_sheet definition

CREATE TABLE `current_sheet` (
  `chat_id` bigint(20) NOT NULL,
  `sheet_id` varchar(36) NOT NULL,
  PRIMARY KEY (`chat_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- budgli.sheet definition

CREATE TABLE `sheet` (
  `sheet_id` varchar(36) NOT NULL,
  `owner_chat_id` bigint(20) NOT NULL,
  `name` varchar(100) NOT NULL,
  `password` varchar(50) NOT NULL,
  PRIMARY KEY (`sheet_id`),
  KEY `sheet_owner_chat_id_IDX` (`owner_chat_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- budgli.payment definition

CREATE TABLE `payment` (
  `payment_id` varchar(36) NOT NULL,
  `sheet_id` varchar(36) NOT NULL,
  `category_id` varchar(36) NOT NULL,
  `amount` bigint(20) NOT NULL,
  `comment` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`payment_id`),
  KEY `payment_sheet_id_IDX` (`sheet_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- budgli.category definition

CREATE TABLE `category` (
  `category_id` varchar(36) NOT NULL,
  `sheet_id` varchar(36) DEFAULT NULL,
  `name` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`category_id`),
  KEY `category_sheet_id_IDX` (`sheet_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;