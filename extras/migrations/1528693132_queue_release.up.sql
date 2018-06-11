-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `ks-installer`.`errors` 
DROP FOREIGN KEY `fk_request_id`;

CREATE TABLE IF NOT EXISTS `ks-installer`.`jobs` (
  `id` VARCHAR(36) NOT NULL,
  `requested_by` VARCHAR(36) NOT NULL,
  `action` TINYINT(1) NOT NULL,
  `state` TINYINT(1) NOT NULL,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `fk_jobs_requested_by_idx` (`requested_by` ASC),
  CONSTRAINT `fk_jobs_requested_by`
    FOREIGN KEY (`requested_by`)
    REFERENCES `ks-installer`.`requests` (`id`)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8;

ALTER TABLE `ks-installer`.`errors` 
DROP COLUMN `link`,
DROP COLUMN `source`,
DROP COLUMN `http_status`,
CHANGE COLUMN `request_id` `request_id` VARCHAR(36) NULL DEFAULT NULL ,
ADD COLUMN `job_id` VARCHAR(36) NULL DEFAULT NULL AFTER `request_id`,
ADD INDEX `fk_errors_request_id_idx` (`request_id` ASC),
ADD INDEX `fk_errors_job_id_idx` (`job_id` ASC),
DROP INDEX `request_id_UNIQUE` ;

ALTER TABLE `ks-installer`.`errors` 
ADD CONSTRAINT `fk_errors_request_id`
  FOREIGN KEY (`request_id`)
  REFERENCES `ks-installer`.`requests` (`id`)
  ON DELETE CASCADE
  ON UPDATE CASCADE,
ADD CONSTRAINT `fk_errors_job_id`
  FOREIGN KEY (`job_id`)
  REFERENCES `ks-installer`.`jobs` (`id`)
  ON DELETE CASCADE
  ON UPDATE CASCADE;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
