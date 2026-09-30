CREATE DATABASE IF NOT EXISTS maz_suplementos CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE maz_suplementos;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  full_name VARCHAR(120) NOT NULL,
  role ENUM('ADMINISTRADOR','CAPTURISTA','AUDITOR') NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_users_username (username)
);

CREATE TABLE IF NOT EXISTS categories (
  id BIGINT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100) NOT NULL, description VARCHAR(500) NOT NULL DEFAULT '', active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_categories_name (name)
);
CREATE TABLE IF NOT EXISTS goals (
  id BIGINT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100) NOT NULL, description VARCHAR(500) NOT NULL DEFAULT '', active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_goals_name (name)
);
CREATE TABLE IF NOT EXISTS ingredients (
  id BIGINT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(120) NOT NULL, description VARCHAR(500) NOT NULL DEFAULT '', active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ingredients_name (name)
);

CREATE TABLE IF NOT EXISTS supplements (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(140) NOT NULL, brand VARCHAR(100) NOT NULL, description TEXT NOT NULL,
  price DECIMAL(12,2) NOT NULL DEFAULT 0 CHECK (price >= 0), stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
  presentation VARCHAR(100) NOT NULL DEFAULT '', flavor VARCHAR(100) NULL, weight VARCHAR(100) NOT NULL DEFAULT '', image_url VARCHAR(2048) NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_supplements_name (name), KEY idx_supplements_brand (brand), KEY idx_supplements_active_stock (active,stock)
);

CREATE TABLE IF NOT EXISTS supplement_categories (
  supplement_id BIGINT NOT NULL, category_id BIGINT NOT NULL, PRIMARY KEY (supplement_id,category_id), KEY idx_sc_category (category_id),
  CONSTRAINT fk_sc_supplement FOREIGN KEY (supplement_id) REFERENCES supplements(id) ON DELETE CASCADE,
  CONSTRAINT fk_sc_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
);
CREATE TABLE IF NOT EXISTS supplement_goals (
  supplement_id BIGINT NOT NULL, goal_id BIGINT NOT NULL, PRIMARY KEY (supplement_id,goal_id), KEY idx_sg_goal (goal_id),
  CONSTRAINT fk_sg_supplement FOREIGN KEY (supplement_id) REFERENCES supplements(id) ON DELETE CASCADE,
  CONSTRAINT fk_sg_goal FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE RESTRICT
);
CREATE TABLE IF NOT EXISTS supplement_ingredients (
  supplement_id BIGINT NOT NULL, ingredient_id BIGINT NOT NULL, PRIMARY KEY (supplement_id,ingredient_id), KEY idx_si_ingredient (ingredient_id),
  CONSTRAINT fk_si_supplement FOREIGN KEY (supplement_id) REFERENCES supplements(id) ON DELETE CASCADE,
  CONSTRAINT fk_si_ingredient FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS orders (
  id CHAR(32) PRIMARY KEY, order_number VARCHAR(24) NOT NULL, customer_name VARCHAR(120) NOT NULL, customer_phone VARCHAR(30) NOT NULL,
  customer_email VARCHAR(160) NULL, status ENUM('PENDIENTE','CONFIRMADO','PREPARANDO','COMPLETADO','CANCELADO') NOT NULL DEFAULT 'PENDIENTE',
  subtotal DECIMAL(12,2) NOT NULL CHECK (subtotal >= 0), total DECIMAL(12,2) NOT NULL CHECK (total >= 0),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_orders_number (order_number), KEY idx_orders_status_created (status,created_at)
);
CREATE TABLE IF NOT EXISTS order_items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT, order_id CHAR(32) NOT NULL, supplement_id BIGINT NOT NULL,
  product_name_snapshot VARCHAR(140) NOT NULL, price_snapshot DECIMAL(12,2) NOT NULL CHECK (price_snapshot >= 0), quantity INT NOT NULL CHECK (quantity > 0), subtotal DECIMAL(12,2) NOT NULL CHECK (subtotal >= 0),
  UNIQUE KEY uk_order_product (order_id,supplement_id), KEY idx_order_items_supplement (supplement_id),
  CONSTRAINT fk_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  CONSTRAINT fk_items_supplement FOREIGN KEY (supplement_id) REFERENCES supplements(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS sessions (
  token_hash CHAR(64) PRIMARY KEY, csrf_hash CHAR(64) NOT NULL, user_id BIGINT NOT NULL, expires_at TIMESTAMP NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_sessions_user (user_id), KEY idx_sessions_expiry (expires_at),
  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
