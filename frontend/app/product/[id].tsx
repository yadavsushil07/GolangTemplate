import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Image,
  ActivityIndicator,
  Alert,
  useWindowDimensions,
} from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { Button } from '@/components/Button';
import { getProduct } from '@/services/api';
import { useCart } from '@/hooks/useCart';

interface Product {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  image_url: string;
  stock: number;
}

export default function ProductDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { width } = useWindowDimensions();
  const { add } = useCart();
  const isWide = width >= 768;

  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const res = await getProduct(Number(id));
        setProduct(res.data);
      } catch {
        Alert.alert('Error', 'Product not found');
        router.back();
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  const handleAdd = async () => {
    if (!product) return;
    setAdding(true);
    try {
      await add(product.id);
      Alert.alert('Added!', `${product.name} has been added to your cart.`);
    } finally {
      setAdding(false);
    }
  };

  if (loading) {
    return <View style={styles.center}><ActivityIndicator color={Colors.primary} size="large" /></View>;
  }

  if (!product) return null;

  return (
    <ScrollView style={styles.container} contentContainerStyle={[styles.content, isWide && styles.contentWide]}>
      {product.image_url ? (
        <Image source={{ uri: product.image_url }} style={[styles.image, isWide && styles.imageWide]} resizeMode="cover" />
      ) : (
        <View style={[styles.imagePlaceholder, isWide && styles.imageWide]}>
          <Text style={styles.emoji}>📦</Text>
        </View>
      )}
      <View style={[styles.info, isWide && styles.infoWide]}>
        <Text style={styles.name}>{product.name}</Text>
        <Text style={styles.price}>${(product.price_cents / 100).toFixed(2)}</Text>
        <Text style={styles.description}>{product.description}</Text>

        {product.stock > 0 ? (
          <>
            <Text style={styles.stock}>{product.stock} in stock</Text>
            <Button title="Add to Cart" onPress={handleAdd} loading={adding} style={styles.btn} />
            <Button title="Go to Cart" onPress={() => router.push('/(tabs)/cart')} variant="outline" style={styles.btn} />
          </>
        ) : (
          <Text style={styles.outOfStock}>Currently out of stock</Text>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: Colors.background },
  content: {},
  contentWide: { flexDirection: 'row' },
  image: { width: '100%', height: 300 },
  imageWide: { width: '45%', height: 400 },
  imagePlaceholder: {
    width: '100%',
    height: 300,
    backgroundColor: Colors.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  emoji: { fontSize: 60 },
  info: { padding: 24, gap: 12 },
  infoWide: { flex: 1 },
  name: { color: Colors.text, fontSize: 24, fontWeight: '800' },
  price: { color: Colors.primary, fontSize: 28, fontWeight: '800' },
  description: { color: Colors.muted, fontSize: 15, lineHeight: 22 },
  stock: { color: Colors.success, fontSize: 13, fontWeight: '600' },
  outOfStock: { color: Colors.error, fontSize: 15, fontWeight: '700' },
  btn: { marginTop: 4 },
});
