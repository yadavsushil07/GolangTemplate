import React, { useEffect, useState, useCallback } from 'react';
import { View, Text, StyleSheet, FlatList, ActivityIndicator, RefreshControl } from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { Button } from '@/components/Button';
import { listMyOrders } from '@/services/api';
import { useAuth } from '@/hooks/useAuth';

interface Order {
  id: number;
  total_cents: number;
  status: string;
  shipping_name: string;
  created_at: string;
}

export default function OrdersScreen() {
  const router = useRouter();
  const { user } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetch = useCallback(async () => {
    if (!user) { setLoading(false); return; }
    try {
      const res = await listMyOrders();
      setOrders(res.data || []);
    } catch {
      setOrders([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [user]);

  useEffect(() => { fetch(); }, [fetch]);

  if (!user) {
    return (
      <View style={styles.center}>
        <Text style={styles.empty}>Sign in to view your orders.</Text>
        <Button title="Sign In" onPress={() => router.push('/(tabs)/profile')} style={{ marginTop: 16 }} />
      </View>
    );
  }

  if (loading) {
    return <View style={styles.center}><ActivityIndicator color={Colors.primary} size="large" /></View>;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>My Orders</Text>
      <FlatList
        data={orders}
        keyExtractor={(o) => String(o.id)}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={() => { setRefreshing(true); fetch(); }} tintColor={Colors.primary} />
        }
        renderItem={({ item }) => (
          <View style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.orderId}>Order #{item.id}</Text>
              <View style={[styles.badge, { backgroundColor: statusColor(item.status) + '33' }]}>
                <Text style={[styles.badgeText, { color: statusColor(item.status) }]}>{item.status}</Text>
              </View>
            </View>
            <Text style={styles.meta}>{new Date(item.created_at).toLocaleDateString()}</Text>
            <Text style={styles.total}>${(item.total_cents / 100).toFixed(2)}</Text>
          </View>
        )}
        ListEmptyComponent={<Text style={styles.empty}>No orders yet.</Text>}
      />
    </View>
  );
}

function statusColor(status: string) {
  switch (status) {
    case 'placed': return Colors.primary;
    case 'shipped': return '#f59e0b';
    case 'delivered': return Colors.success;
    case 'cancelled': return Colors.error;
    default: return Colors.muted;
  }
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: Colors.background },
  title: { color: Colors.text, fontSize: 22, fontWeight: '800', padding: 20, paddingBottom: 8 },
  list: { padding: 16, gap: 12 },
  card: {
    backgroundColor: Colors.surface,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: Colors.border,
    padding: 16,
    gap: 6,
  },
  row: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  orderId: { color: Colors.text, fontSize: 16, fontWeight: '700' },
  badge: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 999 },
  badgeText: { fontSize: 12, fontWeight: '700', textTransform: 'capitalize' },
  meta: { color: Colors.muted, fontSize: 13 },
  total: { color: Colors.primary, fontSize: 18, fontWeight: '800', marginTop: 4 },
  empty: { color: Colors.muted, fontSize: 16, textAlign: 'center' },
});
