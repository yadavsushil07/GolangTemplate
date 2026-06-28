import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  ScrollView,
  StyleSheet,
  Switch,
  Alert,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { Button } from '@/components/Button';
import { CouponInput } from '@/components/CouponInput';
import { checkout } from '@/services/api';
import { useCart } from '@/hooks/useCart';
import { useAuth } from '@/hooks/useAuth';
import * as SecureStore from 'expo-secure-store';

export default function CheckoutScreen() {
  const router = useRouter();
  const { cart, fetchCart } = useCart();
  const { token } = useAuth();

  const [shippingName, setShippingName] = useState('');
  const [shippingAddress, setShippingAddress] = useState('');
  const [customizationNote, setCustomizationNote] = useState('');
  const [useRazorpay, setUseRazorpay] = useState(false);
  const [discountCents, setDiscountCents] = useState(0);
  const [couponCode, setCouponCode] = useState('');
  const [loading, setLoading] = useState(false);

  if (!token) {
    return (
      <View style={styles.center}>
        <Text style={styles.loginText}>Please log in to proceed to checkout.</Text>
        <Button title="Go to Profile" onPress={() => router.push('/(tabs)/profile')} />
      </View>
    );
  }

  const subtotal = cart?.total_cents ?? 0;
  const finalTotal = Math.max(0, subtotal - discountCents);

  const handleCheckout = async () => {
    if (!shippingName || !shippingAddress) {
      Alert.alert('Missing info', 'Please fill in your name and address.');
      return;
    }

    const sessionID = await SecureStore.getItemAsync('session_id') ?? '';
    setLoading(true);
    try {
      const res = await checkout({
        shipping_name: shippingName,
        shipping_address: shippingAddress,
        customization_note: customizationNote,
        session_id: sessionID,
        payment_method: useRazorpay ? 'razorpay' : 'cod',
        coupon_code: couponCode || undefined,
      }) as any;

      await fetchCart();

      const order = res.data;
      if (useRazorpay) {
        // TODO: integrate Razorpay React Native SDK with order.id and finalTotal
        Alert.alert('Razorpay', `Order #${order.id} created. Integrate RN Razorpay SDK here.`);
      } else {
        Alert.alert('Order placed!', `Order #${order.id} confirmed. Pay on delivery.`, [
          { text: 'View Orders', onPress: () => router.push('/(tabs)/orders') },
        ]);
      }
    } catch (err: any) {
      Alert.alert('Error', err?.response?.data?.error ?? 'Checkout failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.heading}>Checkout</Text>

      {/* Shipping details */}
      <View style={styles.section}>
        <Text style={styles.sectionLabel}>Shipping Details</Text>
        <TextInput
          style={styles.input}
          value={shippingName}
          onChangeText={setShippingName}
          placeholder="Full name"
          placeholderTextColor={Colors.muted}
        />
        <TextInput
          style={[styles.input, styles.multiline]}
          value={shippingAddress}
          onChangeText={setShippingAddress}
          placeholder="Full address with pin code"
          placeholderTextColor={Colors.muted}
          multiline
          numberOfLines={3}
        />
      </View>

      {/* Customization note */}
      <View style={styles.section}>
        <Text style={styles.sectionLabel}>Customization Note (optional)</Text>
        <TextInput
          style={[styles.input, styles.multiline]}
          value={customizationNote}
          onChangeText={setCustomizationNote}
          placeholder="E.g. embroidery name, special instructions..."
          placeholderTextColor={Colors.muted}
          multiline
          numberOfLines={3}
        />
      </View>

      {/* Coupon */}
      <View style={styles.section}>
        <CouponInput
          orderTotalCents={subtotal}
          onDiscountApplied={(cents, code) => {
            setDiscountCents(cents);
            setCouponCode(code);
          }}
        />
      </View>

      {/* Payment toggle */}
      <View style={styles.section}>
        <Text style={styles.sectionLabel}>Payment Method</Text>
        <View style={styles.toggle}>
          <Text style={styles.toggleLabel}>{useRazorpay ? 'Online (Razorpay)' : 'Cash on Delivery'}</Text>
          <Switch
            value={useRazorpay}
            onValueChange={setUseRazorpay}
            trackColor={{ false: Colors.border, true: Colors.primary }}
            thumbColor={useRazorpay ? Colors.primaryDark : Colors.muted}
          />
        </View>
      </View>

      {/* Order summary */}
      <View style={styles.section}>
        <Text style={styles.sectionLabel}>Order Summary</Text>
        {cart?.items?.map((item: any) => (
          <View key={item.id} style={styles.summaryRow}>
            <Text style={styles.summaryItem} numberOfLines={1}>
              {item.product?.name}
              {item.variant ? ` (${item.variant.size}${item.variant.color ? ` · ${item.variant.color}` : ''})` : ''}
              {' '}×{item.quantity}
            </Text>
            <Text style={styles.summaryPrice}>
              ₹{((item.variant?.price_cents ?? item.product?.price_cents ?? 0) * item.quantity / 100).toLocaleString('en-IN')}
            </Text>
          </View>
        ))}
        {discountCents > 0 && (
          <View style={styles.summaryRow}>
            <Text style={[styles.summaryItem, styles.discountText]}>Discount ({couponCode})</Text>
            <Text style={[styles.summaryPrice, styles.discountText]}>-₹{(discountCents / 100).toLocaleString('en-IN')}</Text>
          </View>
        )}
        <View style={[styles.summaryRow, styles.totalRow]}>
          <Text style={styles.totalLabel}>Total</Text>
          <Text style={styles.totalValue}>₹{(finalTotal / 100).toLocaleString('en-IN')}</Text>
        </View>
      </View>

      <Button
        title={loading ? 'Placing order...' : 'Place Order'}
        onPress={handleCheckout}
        disabled={loading}
        style={styles.placeBtn}
      />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  content: { padding: 20, paddingBottom: 48, gap: 4 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 16, padding: 24, backgroundColor: Colors.background },
  loginText: { color: Colors.muted, fontSize: 16, textAlign: 'center' },
  heading: { color: Colors.text, fontSize: 26, fontWeight: '800', marginBottom: 8 },
  section: {
    backgroundColor: Colors.surface,
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    gap: 10,
  },
  sectionLabel: { color: Colors.muted, fontSize: 13, fontWeight: '600', marginBottom: 4 },
  input: {
    backgroundColor: Colors.background,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 10,
    padding: 12,
    color: Colors.text,
    fontSize: 15,
  },
  multiline: { minHeight: 80, textAlignVertical: 'top' },
  toggle: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  toggleLabel: { color: Colors.text, fontSize: 15, fontWeight: '600' },
  summaryRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', gap: 8 },
  summaryItem: { flex: 1, color: Colors.text, fontSize: 14 },
  summaryPrice: { color: Colors.text, fontSize: 14, fontWeight: '600' },
  discountText: { color: Colors.success },
  totalRow: { marginTop: 8, paddingTop: 12, borderTopWidth: 1, borderColor: Colors.border },
  totalLabel: { color: Colors.text, fontSize: 17, fontWeight: '700' },
  totalValue: { color: Colors.primary, fontSize: 20, fontWeight: '800' },
  placeBtn: { minHeight: 54 },
});
