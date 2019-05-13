-- MySQL Script generated by MySQL Workbench
-- Tue Nov 14 14:38:54 2017
-- Model: New Model    Version: 1.0
-- MySQL Workbench Forward Engineering

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL,ALLOW_INVALID_DATES';

-- -----------------------------------------------------
-- Schema open_census
-- -----------------------------------------------------

-- -----------------------------------------------------
-- Schema open_census
-- -----------------------------------------------------
CREATE SCHEMA IF NOT EXISTS `open_census` DEFAULT CHARACTER SET utf8 ;
USE `open_census` ;

-- -----------------------------------------------------
-- Table `open_census`.`products`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `open_census`.`products` (
  `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(63) NOT NULL,
  `price` INT NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`))
ENGINE = InnoDB;

INSERT INTO `open_census`.`products`(name, price) values ('tra xanh', 30000);
INSERT INTO `open_census`.`products`(name, price) values ('banh bao', 10000);

SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;

