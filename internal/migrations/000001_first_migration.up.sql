
-- MIGRATION: 0001_init.up.sql
-- PURPOSE: Initialize schema for users, roles, categories,
--          products, relations and product history.

-- USERS & ROLES
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role_id INT REFERENCES roles(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- CATEGORIES
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PRODUCTS
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Many-to-many relation: product_category
CREATE TABLE product_category (
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

-- PRODUCT HISTORY
CREATE TABLE product_history (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price NUMERIC(12,2) NOT NULL,
    stock INT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- INDEXES
CREATE INDEX idx_products_name ON products USING gin (to_tsvector('simple', name));
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_stock ON products(stock);

-- TRIGGER: Record product changes automatically
CREATE OR REPLACE FUNCTION fn_product_history() RETURNS trigger AS $$
BEGIN
    IF (OLD.price IS DISTINCT FROM NEW.price OR OLD.stock IS DISTINCT FROM NEW.stock) THEN
        INSERT INTO product_history (product_id, price, stock, changed_at)
        VALUES (NEW.id, NEW.price, NEW.stock, now());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_product_history
AFTER UPDATE ON products
FOR EACH ROW
EXECUTE PROCEDURE fn_product_history();