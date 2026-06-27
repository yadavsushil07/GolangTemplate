import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { Colors } from '@/constants/colors';
import { Ionicons } from '@expo/vector-icons';

interface Props {
  item: {
    product_id: number;
    quantity: number;
    product: {
      name: string;
      price_cents: number;
      image_url: string;
    };
  };
  onRemove: (productId: number) => void;
}

export function CartItemRow({ item, onRemove }: Props) {
  const lineCents = item.product.price_cents * item.quantity;

  return (
    <View style={styles.row}>
      <View style={styles.info}>
        <Text style={styles.name}>{item.product.name}</Text>
        <Text style={styles.qty}>Qty: {item.quantity}</Text>
      </View>
      <Text style={styles.price}>${(lineCents / 100).toFixed(2)}</Text>
      <TouchableOpacity onPress={() => onRemove(item.product_id)} style={styles.remove}>
        <Ionicons name="trash-outline" size={20} color={Colors.error} />
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderColor: Colors.border,
    gap: 12,
  },
  info: {
    flex: 1,
    gap: 4,
  },
  name: {
    color: Colors.text,
    fontSize: 15,
    fontWeight: '600',
  },
  qty: {
    color: Colors.muted,
    fontSize: 13,
  },
  price: {
    color: Colors.primary,
    fontSize: 15,
    fontWeight: '700',
  },
  remove: {
    padding: 4,
  },
});
