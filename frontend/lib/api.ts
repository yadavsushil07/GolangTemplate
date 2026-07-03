import type {
  Product,
  Category,
  CartSummary,
  Order,
  User,
} from "./types";

const API_BASE =
  (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080") + "/api";

const TOKEN_KEY = "sby_token";
const ROLE_KEY = "sby_role";
const SESSION_KEY = "sby_session";

// ---- local persistence (browser only) ----
const isBrowser = typeof window !== "undefined";

export function getToken(): string | null {
  if (!isBrowser) return null;
  return localStorage.getItem(TOKEN_KEY) ?? sessionStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string, remember: boolean) {
  if (!isBrowser) return;
  if (remember) {
    localStorage.setItem(TOKEN_KEY, token);
    sessionStorage.removeItem(TOKEN_KEY);
  } else {
    sessionStorage.setItem(TOKEN_KEY, token);
    localStorage.removeItem(TOKEN_KEY);
  }
}

export function getRole(): string | null {
  if (!isBrowser) return null;
  return localStorage.getItem(ROLE_KEY) ?? sessionStorage.getItem(ROLE_KEY);
}

export function setRole(role: string, remember: boolean) {
  if (!isBrowser) return;
  if (remember) {
    localStorage.setItem(ROLE_KEY, role);
    sessionStorage.removeItem(ROLE_KEY);
  } else {
    sessionStorage.setItem(ROLE_KEY, role);
    localStorage.removeItem(ROLE_KEY);
  }
}

export function logout() {
  if (!isBrowser) return;
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(ROLE_KEY);
  sessionStorage.removeItem(TOKEN_KEY);
  sessionStorage.removeItem(ROLE_KEY);
}

export function isLoggedIn(): boolean {
  return !!getToken() && !isTokenExpired();
}

export function isTokenExpired(): boolean {
  const token = getToken();
  if (!token) return true;
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    return payload.exp * 1000 < Date.now();
  } catch {
    return true;
  }
}

export function getSessionID(): string {
  if (!isBrowser) return "server";
  let sid = localStorage.getItem(SESSION_KEY);
  if (!sid) {
    sid =
      typeof crypto !== "undefined" && crypto.randomUUID
        ? crypto.randomUUID()
        : "sess-" + Date.now().toString(36);
    localStorage.setItem(SESSION_KEY, sid);
  }
  return sid;
}

// ---- core request helper ----
export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    Accept: "application/json",
    "X-Session-ID": getSessionID(),
    ...(options.headers as Record<string, string>),
  };
  if (options.body) headers["Content-Type"] = "application/json";
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(API_BASE + path, { ...options, headers });
  const text = await res.text();
  const data = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const msg = (data && data.error) || res.statusText || "Request failed";
    throw new ApiError(res.status, msg);
  }
  return data as T;
}

// ---- auth ----
export const requestOTP = (identifier: string) =>
  request<{ message: string; otp?: string }>("/auth/request-otp", {
    method: "POST",
    body: JSON.stringify({ identifier }),
  });

export async function verifyOTP(identifier: string, code: string, rememberMe = false) {
  const res = await request<{ token: string; user: User }>(
    "/auth/verify-otp",
    { method: "POST", body: JSON.stringify({ identifier, code, remember_me: rememberMe }) }
  );
  setToken(res.token, rememberMe);
  if (res.user?.role) setRole(res.user.role, rememberMe);
  return res;
}

// ---- catalogue ----
// Note: the Go API encodes an empty slice as JSON `null`, so list endpoints
// coerce a null response to an empty array.
export const listProducts = async (category?: string) =>
  (await request<Product[] | null>(
    "/products" + (category ? `?category=${encodeURIComponent(category)}` : "")
  )) ?? [];
export const getProductBySlug = (slug: string) =>
  request<Product>(`/products/slug/${encodeURIComponent(slug)}`);
export const getProductById = (id: number) =>
  request<Product>(`/products/${id}`);
export const listCategories = async () =>
  (await request<Category[] | null>("/categories")) ?? [];

// ---- cart ----
export const getCart = () => request<CartSummary>("/cart");
export const addToCart = (
  product_id: number,
  quantity = 1,
  variant_id?: number | null
) =>
  request<CartSummary>("/cart", {
    method: "POST",
    body: JSON.stringify({ product_id, quantity, variant_id }),
  });
export const removeFromCart = (productId: number) =>
  request<CartSummary>(`/cart/${productId}`, { method: "DELETE" });

// ---- coupons ----
export const validateCoupon = (code: string, order_total_cents: number) =>
  request<{ valid: boolean; discount_cents: number; message?: string }>(
    "/coupons/validate",
    { method: "POST", body: JSON.stringify({ code, order_total_cents }) }
  );

// ---- orders / checkout (auth required) ----
export interface CheckoutPayload {
  shipping_name: string;
  shipping_address: string;
  payment_method: string;
  coupon_code?: string;
  customization_note?: string;
  session_id: string;
}
export const checkout = (payload: CheckoutPayload) =>
  request<Order>("/checkout", {
    method: "POST",
    body: JSON.stringify(payload),
  });
export const listMyOrders = async () =>
  (await request<Order[] | null>("/orders")) ?? [];

// ---- vendor / admin (role required) ----

export interface VendorStats {
  total_revenue_cents: number;
  total_orders: number;
  total_products: number;
  low_stock_products: number;
  orders_by_status: Record<string, number>;
  monthly_revenue: Array<{
    year: number;
    month: number;
    revenue_cents: number;
    order_count: number;
  }>;
}

export const vendorGetStats = () => request<VendorStats>("/vendor/stats");

export const vendorListProducts = async () =>
  (await request<Product[] | null>("/vendor/products")) ?? [];
export const vendorCreateProduct = (data: {
  name: string;
  description?: string;
  price_cents: number;
  image_url?: string;
  stock: number;
}) =>
  request<Product>("/vendor/products", {
    method: "POST",
    body: JSON.stringify(data),
  });
export const vendorUpdateProduct = (id: number, data: Record<string, unknown>) =>
  request<Product>(`/vendor/products/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
export const vendorDeactivateProduct = (id: number) =>
  request<{ message: string }>(`/vendor/products/${id}`, { method: "DELETE" });
export const vendorListOrders = async () =>
  (await request<Order[] | null>("/vendor/orders")) ?? [];
export const vendorUpdateOrderStatus = (id: number, status: string) =>
  request<Order>(`/vendor/orders/${id}/status`, {
    method: "PUT",
    body: JSON.stringify({ status }),
  });

// ---- vendor categories ----
export const vendorListCategories = async () =>
  (await request<Category[] | null>("/vendor/categories")) ?? [];
export const vendorCreateCategory = (data: { name: string; slug: string; sort_order?: number }) =>
  request<Category>("/vendor/categories", {
    method: "POST",
    body: JSON.stringify(data),
  });
export const vendorDeleteCategory = (id: number) =>
  request<{ message: string }>(`/vendor/categories/${id}`, { method: "DELETE" });

// ---- vendor coupons ----
export interface Coupon {
  id: number;
  code: string;
  type: string;
  value: number;
  min_order_cents: number;
  expires_at?: string;
  is_active: boolean;
  created_at: string;
}
export interface CreateCouponPayload {
  code: string;
  type: "flat" | "percent";
  value: number;
  min_order_cents?: number;
  expires_at?: string;
}
export const vendorListCoupons = async () =>
  (await request<Coupon[] | null>("/vendor/coupons")) ?? [];
export const vendorCreateCoupon = (data: CreateCouponPayload) =>
  request<Coupon>("/vendor/coupons", {
    method: "POST",
    body: JSON.stringify(data),
  });
export const vendorDeactivateCoupon = (id: number) =>
  request<{ message: string }>(`/vendor/coupons/${id}`, { method: "DELETE" });
