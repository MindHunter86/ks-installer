-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `ks-installer`.`macs` 
DROP FOREIGN KEY `fk_macs_host`;

ALTER TABLE `ks-installer`.`macs` 
ADD CONSTRAINT `fk_macs_host`
  FOREIGN KEY (`host`)
  REFERENCES `ks-installer`.`hosts` (`id`)
  ON DELETE RESTRICT
  ON UPDATE RESTRICT;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
