CREATE INDEX IF NOT EXISTS idx_product_name
    ON Product (Name);

CREATE INDEX IF NOT EXISTS idx_sellerproduct_seller
    ON SellerProduct (seller_id);

CREATE INDEX IF NOT EXISTS idx_sellerproduct_product
    ON SellerProduct (product_id);

CREATE INDEX IF NOT EXISTS idx_sellerproduct_ext
    ON SellerProduct (seller_id, external_id);
