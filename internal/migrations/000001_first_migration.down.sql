
-- MIGRATION: 0001_init.down.sql
-- PURPOSE: Rollback schema initialization

DROP TRIGGER IF EXISTS trg_product_history ON products;
DROP FUNCTION IF EXISTS fn_product_history;

DROP TABLE IF EXISTS product_history;
DROP TABLE IF EXISTS product_category;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;