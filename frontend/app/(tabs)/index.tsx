import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  StyleSheet,
  useWindowDimensions,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { ProductCard } from '@/components/ProductCard';
import { listProducts } from '@/services/api';
import { useCart } from '@/hooks/useCart';

interface Product {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  image_url: string;
  stock: number;
}

export default function HomeScreen() {
  const router = useRouter();
  const { width } = useWindowDimensions();
  const { add } = useCart();

  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [addingId, setAddingId] = useState<number | null>(null);

  const numColumns = width >= 1024 ? 3 : width >= 768 ? 2 : 1;

  const fetchProducts = useCallback(async () => {
    try {
      const res = await listProducts();
      setProducts(res.data || []);
    } catch (e) {
      console.error('Failed to fetch products', e);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { fetchProducts(); }, [fetchProducts]);

  const handleAddToCart = useCallback(async (id: number) => {
    setAddingId(id);
    try {
      await add(id);
    } finally {
      setAddingId(null);
    }
  }, [add]);

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={Colors.primary} size="large" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.hero}>
        <Text style={styles.heroTitle}>Discover Our Products</Text>
        <Text style={styles.heroSub}>Premium quality, delivered to your door</Text>
      </View>
      <FlatList
        data={products}
        key={numColumns}
        keyExtractor={(p) => String(p.id)}
        numColumns={numColumns}
        columnWrapperStyle={numColumns > 1 ? styles.row : undefined}
        contentContainerStyle={styles.list}
        renderItem={({ item }) => (
          <ProductCard
            product={item}
            onAddToCart={handleAddToCart}
            onPress={(id) => router.push(`/product/${id}`)}
            adding={addingId === item.id}
          />
        )}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={() => { setRefreshing(true); fetchProducts(); }}
            tintColor={Colors.primary}
          />
        }
        ListEmptyComponent={
          <Text style={styles.empty}>No products available yet.</Text>
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: Colors.background },
  hero: {
    padding: 24,
    paddingBottom: 8,
    backgroundColor: Colors.surface,
    borderBottomWidth: 1,
    borderColor: Colors.border,
  },
  heroTitle: { color: Colors.text, fontSize: 22, fontWeight: '800' },
  heroSub: { color: Colors.muted, fontSize: 14, marginTop: 4 },
  list: { padding: 16 },
  row: { gap: 16 },
  empty: { color: Colors.muted, textAlign: 'center', marginTop: 48, fontSize: 16 },
});
