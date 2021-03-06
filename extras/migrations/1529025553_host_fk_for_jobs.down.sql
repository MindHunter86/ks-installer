-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `ks-installer`.`hosts` 
DROP FOREIGN KEY `fk_hosts_created_by`;

ALTER TABLE `ks-installer`.`hosts` 
DROP COLUMN `created_by`,
CHANGE COLUMN `hostname` `hostname` VARCHAR(45) NOT NULL ,
DROP INDEX `fk_hosts_created_by_idx` ;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
