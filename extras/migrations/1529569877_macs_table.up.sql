-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `ks-installer`.`requests` 
CHANGE COLUMN `size` `size` INT(10) UNSIGNED NOT NULL ,
CHANGE COLUMN `status` `status` SMALLINT(2) UNSIGNED NOT NULL ;

ALTER TABLE `ks-installer`.`jobs` 
CHANGE COLUMN `action` `action` TINYINT(1) UNSIGNED NOT NULL ,
CHANGE COLUMN `state` `state` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 ,
CHANGE COLUMN `is_failed` `is_failed` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 ;

CREATE TABLE IF NOT EXISTS `ks-installer`.`macs` (
  `mac` VARCHAR(17) NOT NULL,
  `host` VARCHAR(36) NULL DEFAULT NULL,
  `jun_number` SMALLINT(5) UNSIGNED NULL DEFAULT NULL,
  `jun_port_name` VARCHAR(16) NULL DEFAULT NULL,
  `jun_vlan` SMALLINT(5) UNSIGNED NULL DEFAULT NULL,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`mac`),
  INDEX `fk_macs_host_idx` (`host` ASC),
  CONSTRAINT `fk_macs_host`
    FOREIGN KEY (`host`)
    REFERENCES `ks-installer`.`hosts` (`id`)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8;

DROP TABLE IF EXISTS `ks-installer`.`ports` ;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
