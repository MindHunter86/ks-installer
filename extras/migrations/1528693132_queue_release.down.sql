-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `ks-installer`.`errors` 
DROP FOREIGN KEY `fk_errors_job_id`,
DROP FOREIGN KEY `fk_errors_request_id`;

ALTER TABLE `ks-installer`.`errors` 
DROP COLUMN `job_id`,
CHANGE COLUMN `request_id` `request_id` VARCHAR(36) NOT NULL ,
ADD COLUMN `http_status` SMALLINT(2) UNSIGNED NOT NULL AFTER `internal_code`,
ADD COLUMN `source` VARCHAR(18) NULL DEFAULT NULL AFTER `http_status`,
ADD COLUMN `link` VARCHAR(32) NOT NULL AFTER `source`,
ADD UNIQUE INDEX `request_id_UNIQUE` (`request_id` ASC),
DROP INDEX `fk_errors_job_id_idx` ,
DROP INDEX `fk_errors_request_id_idx` ;

DROP TABLE IF EXISTS `ks-installer`.`jobs` ;

ALTER TABLE `ks-installer`.`errors` 
ADD CONSTRAINT `fk_request_id`
  FOREIGN KEY (`request_id`)
  REFERENCES `ks-installer`.`requests` (`id`)
  ON DELETE RESTRICT
  ON UPDATE RESTRICT;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
