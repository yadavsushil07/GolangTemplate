import axios from 'axios';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const BASE_URL = process.env.EXPO_PUBLIC_API_URL
  ? `${process.env.EXPO_PUBLIC_API_URL}/api`
  : Platform.OS === 'android'
    ? 'http://10.0.2.2:8080/api'
    : 'http://localhost:8080/api';

const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

const TOKEN_KEY = 'auth_token';
const SESSION_KEY = 'session_id';

export async function getToken(): Promise<string | null> {
  if (Platform.OS === 'web') {
    return localStorage.getItem(TOKEN_KEY);
  }
  return SecureStore.getItemAsync(TOKEN_KEY);
}

export async function setToken(token: string): Promise<void> {
  if (Platform.OS === 'web') {
    localStorage.setItem(TOKEN_KEY, token);
  } else {
    await SecureStore.setItemAsync(TOKEN_KEY, token);
  }
}

export async function clearToken(): Promise<void> {
  if (Platform.OS === 'web') {
    localStorage.removeItem(TOKEN_KEY);
  } else {
    await SecureStore.deleteItemAsync(TOKEN_KEY);
  }
}

export async function getSessionID(): Promise<string> {
  let sid: string | null = null;
  if (Platform.OS === 'web') {
    sid = localStorage.getItem(SESSION_KEY);
  } else {
    sid = await SecureStore.getItemAsync(SESSION_KEY);
  }
  if (!sid) {
    sid = Math.random().toString(36).slice(2) + Date.now().toString(36);
    if (Platform.OS === 'web') {
      localStorage.setItem(SESSION_KEY, sid);
    } else {
      await SecureStore.setItemAsync(SESSION_KEY, sid);
    }
  }
  return sid;
}

api.interceptors.request.use(async (config) => {
  const token = await getToken();
  const sessionID = await getSessionID();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  config.headers['X-Session-ID'] = sessionID;
  return config;
});

// --- Categories ---
export const listCategories = () => api.get('/categories');

// --- Coupons ---
export const validateCoupon = (code: string, order_total_cents: number) =>
  api.post('/coupons/validate', { code, order_total_cents });

// --- Auth ---
export const requestOTP = (identifier: string) =>
  api.post('/auth/request-otp', { identifier });

export const verifyOTP = (identifier: string, code: string) =>
  api.post('/auth/verify-otp', { identifier, code });

// --- Products ---
export const listProducts = () => api.get('/products');
export const getProduct = (id: number) => api.get(`/products/${id}`);

// --- Cart ---
export const getCart = () => api.get('/cart');
export const addToCart = (product_id: number, quantity: number = 1) =>
  api.post('/cart', { product_id, quantity });
export const removeFromCart = (productId: number) =>
  api.delete(`/cart/${productId}`);

// --- Orders ---
export const checkout = (shipping_name: string, shipping_address: string, session_id: string) =>
  api.post('/checkout', { shipping_name, shipping_address, session_id });
export const listMyOrders = () => api.get('/orders');

// --- Razorpay ---
export const createRazorpayOrder = (order_id: number, amount_cents: number) =>
  api.post('/payments/razorpay/create-order', { order_id, amount_cents });
export const verifyRazorpayPayment = (data: object) =>
  api.post('/payments/razorpay/verify', data);

// --- Vendor ---
export const vendorListProducts = () => api.get('/vendor/products');
export const vendorCreateProduct = (data: object) => api.post('/vendor/products', data);
export const vendorUpdateProduct = (id: number, data: object) =>
  api.put(`/vendor/products/${id}`, data);
export const vendorDeactivateProduct = (id: number) =>
  api.delete(`/vendor/products/${id}`);
export const vendorListOrders = () => api.get('/vendor/orders');
export const vendorUpdateOrderStatus = (id: number, status: string) =>
  api.put(`/vendor/orders/${id}/status`, { status });

export default api;
