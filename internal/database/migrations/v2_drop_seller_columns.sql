-- 1. Dropar a tabela SellerProduct
-- DROP TABLE IF EXISTS SellerProduct;
-- 2. Recriar a tabela SellerProduct com o novo schema
CREATE TABLE IF NOT EXISTS SellerProduct (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  seller_id TEXT NOT NULL,
  product_id INTEGER NOT NULL,
  external_id TEXT NOT NULL,
  CONSTRAINT FK_Seller_Id FOREIGN KEY (seller_id) REFERENCES Seller (id),
  CONSTRAINT FK_Product_Id FOREIGN KEY (product_id) REFERENCES Product (Id),
  CONSTRAINT UQ_seller_product UNIQUE (seller_id, product_id)
);