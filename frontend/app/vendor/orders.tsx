import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  ActivityIndicator,
  RefreshControl,
  TouchableOpacity,
  Alert,
  useWindowDimensions,
} from 'react-native';
import { Colors } from '@/constants/colors';
import { vendorListOrders, vendorUpdateOrderStatus } from '@/services/api';

interface Order {
  id: number;
  user_id: number;
  total_cents: number;
  status: string;
  shipping_name: string;
  shipping_address: string;
  created_at: string;
}

const STATUSES = ['placed', 'shipped', 'delivered', 'cancelled'];

export default function VendorOrdersScreen() {
  const { width } = useWindowDimensions();
  const isWide = width >= 768;

  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetch = useCallback(async () => {
    try {
      const res = await vendorListOrders();
      setOrders(res.data || []);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);

  const handleStatusChange = (order: Order) => {
    Alert.alert('Update Status', `Order #${order.id}`, [
      ...STATUSES.map((s) => ({
        text: s.charAt(0).toUpperCase() + s.slice(1),
        onPress: async () => {
          await vendorUpdateOrderStatus(order.id, s);
          fetch();
        },
      })),
      { text: 'Cancel', style: 'cancel' },
    ]);
  };

  if (loading) {
    return <View style={styles.center}><ActivityIndicator color={Colors.primary} size="large" /></View>;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>All Orders ({orders.length})</Text>
      <FlatList
        data={orders}
        keyExtractor={(o) => String(o.id)}
        key={isWide ? 2 : 1}
        numColumns={isWide ? 2 : 1}
        columnWrapperStyle={isWide ? styles.row : undefined}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={() => { setRefreshing(true); fetch(); }} tintColor={Colors.primary} />
        }
        renderItem={({ item }) => (
          <TouchableOpacity
            style={[styles.card, isWide && styles.cardWide]}
            onPress={() => handleStatusChange(item)}
            activeOpacity={0.85}
          >
            <View style={styles.cardHeader}>
              <Text style={styles.orderId}>Order #{item.id}</Text>
              <View style={[styles.badge, { backgroundColor: statusColor(item.status) + '33' }]}>
                <Text style={[styles.badgeText, { color: statusColor(item.status) }]}>{item.status}</Text>
              </View>
            </View>
            <Text style={styles.customer}>{item.shipping_name}</Text>
            <Text style={styles.address} numberOfLines={2}>{item.shipping_address}</Text>
            <View style={styles.footer}>
              <Text style={styles.date}>{new Date(item.created_at).toLocaleDateString()}</Text>
              <Text style={styles.total}>${(item.total_cents / 100).toFixed(2)}</Text>
            </View>
            <Text style={styles.tapHint}>Tap to update status</Text>
          </TouchableOpacity>
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
  title: { color: Colors.text, fontSize: 20, fontWeight: '800', padding: 20, paddingBottom: 8 },
  list: { padding: 16, gap: 12 },
  row: { gap: 12 },
  card: {
    backgroundColor: Colors.surface,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: Colors.border,
    padding: 16,
    gap: 6,
  },
  cardWide: { flex: 1 },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  orderId: { color: Colors.text, fontSize: 16, fontWeight: '700' },
  badge: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 999 },
  badgeText: { fontSize: 12, fontWeight: '700', textTransform: 'capitalize' },
  customer: { color: Colors.text, fontSize: 14, fontWeight: '600' },
  address: { color: Colors.muted, fontSize: 13 },
  footer: { flexDirection: 'row', justifyContent: 'space-between', marginTop: 6 },
  date: { color: Colors.muted, fontSize: 13 },
  total: { color: Colors.primary, fontSize: 16, fontWeight: '800' },
  tapHint: { color: Colors.muted, fontSize: 11, marginTop: 4 },
  empty: { color: Colors.muted, textAlign: 'center', marginTop: 48, fontSize: 16 },
});
