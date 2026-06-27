import React, { useEffect } from 'react';
import { View, Text, StyleSheet, FlatList, ActivityIndicator } from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { CartItemRow } from '@/components/CartItem';
import { Button } from '@/components/Button';
import { useCart } from '@/hooks/useCart';

export default function CartScreen() {
  const router = useRouter();
  const { cart, loading, fetchCart, remove } = useCart();

  useEffect(() => { fetchCart(); }, [fetchCart]);

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={Colors.primary} size="large" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Your Cart</Text>
      {cart.items.length === 0 ? (
        <View style={styles.center}>
          <Text style={styles.empty}>Your cart is empty.</Text>
          <Button title="Browse Products" onPress={() => router.push('/')} style={{ marginTop: 20 }} />
        </View>
      ) : (
        <>
          <FlatList
            data={cart.items}
            keyExtractor={(i) => String(i.product_id)}
            contentContainerStyle={styles.list}
            renderItem={({ item }) => (
              <CartItemRow item={item} onRemove={remove} />
            )}
          />
          <View style={styles.footer}>
            <View style={styles.totalRow}>
              <Text style={styles.totalLabel}>Total</Text>
              <Text style={styles.total}>${(cart.total_cents / 100).toFixed(2)}</Text>
            </View>
            <Button
              title="Proceed to Checkout"
              onPress={() => router.push('/checkout')}
              style={{ marginTop: 12 }}
            />
          </View>
        </>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  title: { color: Colors.text, fontSize: 22, fontWeight: '800', padding: 20, paddingBottom: 8 },
  list: { padding: 16 },
  empty: { color: Colors.muted, fontSize: 16 },
  footer: {
    padding: 20,
    borderTopWidth: 1,
    borderColor: Colors.border,
    backgroundColor: Colors.surface,
  },
  totalRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  totalLabel: { color: Colors.muted, fontSize: 16 },
  total: { color: Colors.text, fontSize: 20, fontWeight: '800' },
});
