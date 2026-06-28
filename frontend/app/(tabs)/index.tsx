import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  FlatList,
  StyleSheet,
  useWindowDimensions,
  RefreshControl,
  ActivityIndicator,
  Text,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { ProductCard } from '@/components/ProductCard';
import { CategoryFilter } from '@/components/CategoryFilter';
import { listProducts, listCategories } from '@/services/api';
import { useCart } from '@/hooks/useCart';

interface Product {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  image_url: string;
  stock: number;
  images: { url: string }[];
}

interface Category {
  id: number;
  name: string;
  slug: string;
}

export default function HomeScreen() {
  const router = useRouter();
  const { width } = useWindowDimensions();
  const { add } = useCart();

  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategory, setSelectedCategory] = useState('');
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [addingId, setAddingId] = useState<number | null>(null);

  const numColumns = width >= 1024 ? 3 : width >= 768 ? 2 : 1;

  const fetchData = useCallback(async () => {
    try {
      const [prodRes, catRes] = await Promise.all([
        listProducts() as any,
        listCategories() as any,
      ]);
      setProducts(prodRes.data || []);
      setCategories(catRes.data || []);
    } catch {
      // no-op
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  const fetchProducts = useCallback(async () => {
    try {
      const res = await listProducts() as any;
      setProducts(res.data || []);
    } finally {
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    (async () => {
      try {
        const res = await listProducts() as any;
        setProducts(res.data || []);
      } catch {
        // no-op
      }
    })();
  }, [selectedCategory]);

  const handleAddToCart = useCallback(async (id: number) => {
    setAddingId(id);
    try {
      await add(id);
    } finally {
      setAddingId(null);
    }
  }, [add]);

  if (loading) {
    return <View style={styles.center}><ActivityIndicator color={Colors.primary} size="large" /></View>;
  }

  return (
    <View style={styles.container}>
      <View style={styles.hero}>
        <Text style={styles.heroTitle}>Handcrafted Elegance</Text>
        <Text style={styles.heroSub}>Celebrating Indian heritage</Text>
      </View>
      <CategoryFilter
        categories={categories}
        selected={selectedCategory}
        onSelect={setSelectedCategory}
      />
      <FlatList
        data={products}
        key={numColumns}
        keyExtractor={(p) => String(p.id)}
        numColumns={numColumns}
        columnWrapperStyle={numColumns > 1 ? styles.row : undefined}
        contentContainerStyle={styles.list}
        renderItem={({ item }) => (
          <ProductCard
            product={{ ...item, image_url: item.images?.[0]?.url || item.image_url }}
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
        ListEmptyComponent={<Text style={styles.empty}>No products available yet.</Text>}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: Colors.background },
  hero: {
    padding: 20,
    paddingBottom: 12,
    backgroundColor: Colors.surface,
  },
  heroTitle: { color: Colors.text, fontSize: 22, fontWeight: '800' },
  heroSub: { color: Colors.muted, fontSize: 13, marginTop: 2 },
  list: { padding: 16 },
  row: { gap: 16 },
  empty: { color: Colors.muted, textAlign: 'center', marginTop: 48, fontSize: 16 },
});
