export interface Category {
  id: number;
  name: string;
  slug: string;
  sort_order: number;
}

export interface ProductVariant {
  id: number;
  product_id: number;
  size: string;
  color: string;
  price_cents: number;
  stock: number;
  sku: string;
  is_active: boolean;
}

export interface ProductImage {
  id: number;
  product_id: number;
  url: string;
  sort_order: number;
}

export interface AttributeValue {
  id: number;
  attribute_id: number;
  value: string;
  sort_order: number;
}

export interface Product {
  id: number;
  name: string;
  slug: string;
  description: string;
  price_cents: number;
  image_url: string;
  stock: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  variants?: ProductVariant[];
  images?: ProductImage[];
  categories?: Category[];
  attribute_values?: AttributeValue[];
}

export interface CartItem {
  id: number;
  session_id: string;
  product_id: number;
  variant_id?: number | null;
  quantity: number;
  product?: Product;
  variant?: ProductVariant;
}

export interface CartSummary {
  items: CartItem[];
  total_cents: number;
}

export interface User {
  id: number;
  identifier: string;
  phone?: string;
  email?: string;
  role: "customer" | "vendor" | "admin";
  created_at: string;
}

export interface Order {
  id: number;
  user_id: number;
  total_cents: number;
  status: string;
  payment_method: string;
  payment_status: string;
  shipping_name: string;
  shipping_address: string;
  created_at: string;
  items?: OrderItem[];
}

export interface OrderItem {
  id: number;
  order_id: number;
  product_id: number;
  variant_id?: number | null;
  quantity: number;
  price_cents: number;
}

export interface AuthResponse {
  token: string;
  role: string;
}

export interface CouponResult {
  valid: boolean;
  discount_cents: number;
  message?: string;
}
