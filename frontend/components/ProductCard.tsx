import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Image, useWindowDimensions } from 'react-native';
import { Colors } from '@/constants/colors';
import { Button } from './Button';

interface Product {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  image_url: string;
  stock: number;
}

interface Props {
  product: Product;
  onAddToCart: (id: number) => void;
  onPress: (id: number) => void;
  adding?: boolean;
}

export function ProductCard({ product, onAddToCart, onPress, adding }: Props) {
  const { width } = useWindowDimensions();
  const isWide = width >= 768;

  return (
    <TouchableOpacity
      style={[styles.card, isWide && styles.cardWide]}
      onPress={() => onPress(product.id)}
      activeOpacity={0.9}
    >
      {product.image_url ? (
        <Image source={{ uri: product.image_url }} style={styles.image} resizeMode="cover" />
      ) : (
        <View style={styles.imagePlaceholder}>
          <Text style={styles.imagePlaceholderText}>📦</Text>
        </View>
      )}
      <View style={styles.body}>
        <Text style={styles.name} numberOfLines={2}>{product.name}</Text>
        <Text style={styles.desc} numberOfLines={2}>{product.description}</Text>
        <View style={styles.footer}>
          <Text style={styles.price}>${(product.price_cents / 100).toFixed(2)}</Text>
          {product.stock > 0 ? (
            <Button
              title="Add to cart"
              onPress={() => onAddToCart(product.id)}
              loading={adding}
              style={styles.btn}
            />
          ) : (
            <Text style={styles.outOfStock}>Out of stock</Text>
          )}
        </View>
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: Colors.surface,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: Colors.border,
    overflow: 'hidden',
    marginBottom: 16,
  },
  cardWide: {
    flex: 1,
    margin: 8,
  },
  image: {
    width: '100%',
    height: 180,
  },
  imagePlaceholder: {
    width: '100%',
    height: 140,
    backgroundColor: Colors.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  imagePlaceholderText: {
    fontSize: 40,
  },
  body: {
    padding: 16,
    gap: 8,
  },
  name: {
    color: Colors.text,
    fontSize: 16,
    fontWeight: '700',
  },
  desc: {
    color: Colors.muted,
    fontSize: 13,
  },
  footer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 8,
  },
  price: {
    color: Colors.primary,
    fontSize: 18,
    fontWeight: '800',
  },
  btn: {
    paddingVertical: 8,
    paddingHorizontal: 14,
    minHeight: 36,
  },
  outOfStock: {
    color: Colors.error,
    fontSize: 13,
    fontWeight: '600',
  },
});
