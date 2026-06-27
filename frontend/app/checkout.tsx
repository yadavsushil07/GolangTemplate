import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  StyleSheet,
  ScrollView,
  Alert,
  KeyboardAvoidingView,
  Platform,
  useWindowDimensions,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { Button } from '@/components/Button';
import { CartItemRow } from '@/components/CartItem';
import { checkout, getSessionID } from '@/services/api';
import { useCart } from '@/hooks/useCart';
import { useAuth } from '@/hooks/useAuth';

export default function CheckoutScreen() {
  const router = useRouter();
  const { user } = useAuth();
  const { cart, fetchCart } = useCart();
  const { width } = useWindowDimensions();
  const isWide = width >= 768;

  const [name, setName] = useState('');
  const [address, setAddress] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => { fetchCart(); }, [fetchCart]);

  const handleCheckout = async () => {
    if (!user) {
      Alert.alert('Sign in required', 'Please sign in to place an order.');
      router.push('/(tabs)/profile');
      return;
    }
    if (!name.trim() || !address.trim()) {
      Alert.alert('Error', 'Please fill in all fields.');
      return;
    }
    if (cart.items.length === 0) {
      Alert.alert('Error', 'Your cart is empty.');
      return;
    }

    setSubmitting(true);
    try {
      const sid = await getSessionID();
      const res = await checkout(name.trim(), address.trim(), sid);
      Alert.alert('Order Placed!', `Your order #${res.data.id} has been placed successfully.`, [
        { text: 'View Orders', onPress: () => router.replace('/(tabs)/orders') },
      ]);
    } catch (e: any) {
      Alert.alert('Error', e?.response?.data?.error || 'Failed to place order.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView contentContainerStyle={[styles.content, isWide && styles.contentWide]}>
        <View style={[styles.form, isWide && styles.formWide]}>
          <Text style={styles.sectionTitle}>Shipping Details</Text>
          <View style={styles.card}>
            <Text style={styles.label}>Full Name</Text>
            <TextInput
              style={styles.input}
              value={name}
              onChangeText={setName}
              placeholder="John Doe"
              placeholderTextColor={Colors.muted}
            />
            <Text style={styles.label}>Delivery Address</Text>
            <TextInput
              style={[styles.input, styles.textarea]}
              value={address}
              onChangeText={setAddress}
              placeholder="123 Main St, City, Country"
              placeholderTextColor={Colors.muted}
              multiline
              numberOfLines={3}
            />
          </View>
          <Button
            title={`Pay $${(cart.total_cents / 100).toFixed(2)}`}
            onPress={handleCheckout}
            loading={submitting}
            disabled={cart.items.length === 0}
          />
        </View>

        <View style={[styles.summary, isWide && styles.summaryWide]}>
          <Text style={styles.sectionTitle}>Order Summary</Text>
          <View style={styles.card}>
            {cart.items.map((item) => (
              <CartItemRow key={item.product_id} item={item} onRemove={() => {}} />
            ))}
            {cart.items.length === 0 && (
              <Text style={styles.empty}>Cart is empty</Text>
            )}
            <View style={styles.totalRow}>
              <Text style={styles.totalLabel}>Total</Text>
              <Text style={styles.total}>${(cart.total_cents / 100).toFixed(2)}</Text>
            </View>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  content: { padding: 20, gap: 20 },
  contentWide: { flexDirection: 'row', alignItems: 'flex-start' },
  form: { flex: 1, gap: 16 },
  formWide: { flex: 1, marginRight: 16 },
  summary: { gap: 16 },
  summaryWide: { flex: 1 },
  sectionTitle: { color: Colors.text, fontSize: 18, fontWeight: '700' },
  card: {
    backgroundColor: Colors.surface,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: Colors.border,
    padding: 16,
    gap: 10,
  },
  label: { color: Colors.muted, fontSize: 13, fontWeight: '600' },
  input: {
    backgroundColor: Colors.background,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 10,
    padding: 12,
    color: Colors.text,
    fontSize: 15,
  },
  textarea: { height: 80, textAlignVertical: 'top' },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 8,
    paddingTop: 12,
    borderTopWidth: 1,
    borderColor: Colors.border,
  },
  totalLabel: { color: Colors.muted, fontSize: 16 },
  total: { color: Colors.text, fontSize: 20, fontWeight: '800' },
  empty: { color: Colors.muted, fontSize: 14, textAlign: 'center', padding: 16 },
});
