import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, ScrollView } from 'react-native';
import { Colors } from '@/constants/colors';

interface Variant {
  id: number;
  size: string;
  color: string;
  price_cents: number;
  stock: number;
  is_active: boolean;
}

interface Props {
  variants: Variant[];
  selectedVariantId: number | null;
  onSelect: (variant: Variant) => void;
}

export function VariantSelector({ variants, selectedVariantId, onSelect }: Props) {
  const sizes = [...new Set(variants.map((v) => v.size))];
  const colors = [...new Set(variants.filter((v) => v.color).map((v) => v.color))];
  const hasColors = colors.length > 0;

  const selectedVariant = variants.find((v) => v.id === selectedVariantId) ?? null;

  return (
    <View style={styles.container}>
      {/* Size selector */}
      <Text style={styles.label}>Size</Text>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.row}>
        {sizes.map((size) => {
          const variantsForSize = variants.filter((v) => v.size === size);
          const isAvailable = variantsForSize.some((v) => v.stock > 0 && v.is_active);
          const isSelected = variantsForSize.some((v) => v.id === selectedVariantId);

          return (
            <TouchableOpacity
              key={size}
              style={[
                styles.sizeChip,
                isSelected && styles.sizeChipSelected,
                !isAvailable && styles.sizeChipDisabled,
              ]}
              onPress={() => {
                if (!isAvailable) return;
                const match = variantsForSize.find((v) => v.is_active && v.stock > 0);
                if (match) onSelect(match);
              }}
              disabled={!isAvailable}
            >
              <Text style={[styles.sizeText, isSelected && styles.sizeTextSelected, !isAvailable && styles.sizeTextDisabled]}>
                {size}
              </Text>
            </TouchableOpacity>
          );
        })}
      </ScrollView>

      {/* Color / fabric selector (if present) */}
      {hasColors && (
        <>
          <Text style={[styles.label, { marginTop: 12 }]}>Color / Fabric</Text>
          <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.row}>
            {colors.map((color) => {
              const variantsForColor = variants.filter((v) => v.color === color);
              const isSelected = variantsForColor.some((v) => v.id === selectedVariantId);
              const isAvailable = variantsForColor.some((v) => v.stock > 0 && v.is_active);

              return (
                <TouchableOpacity
                  key={color}
                  style={[styles.colorChip, isSelected && styles.colorChipSelected, !isAvailable && styles.sizeChipDisabled]}
                  onPress={() => {
                    if (!isAvailable) return;
                    // Match currently selected size in the new color
                    const currentSize = selectedVariant?.size;
                    const match = variants.find(
                      (v) => v.color === color && v.size === currentSize && v.is_active && v.stock > 0
                    ) ?? variantsForColor.find((v) => v.is_active && v.stock > 0);
                    if (match) onSelect(match);
                  }}
                  disabled={!isAvailable}
                >
                  <Text style={[styles.colorText, isSelected && styles.colorTextSelected]}>{color}</Text>
                </TouchableOpacity>
              );
            })}
          </ScrollView>
        </>
      )}

      {selectedVariant && (
        <View style={styles.priceRow}>
          <Text style={styles.selectedPrice}>
            ₹{(selectedVariant.price_cents / 100).toLocaleString('en-IN')}
          </Text>
          <Text style={styles.stockInfo}>
            {selectedVariant.stock > 0 ? `${selectedVariant.stock} in stock` : 'Out of stock'}
          </Text>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 8 },
  label: { color: Colors.muted, fontSize: 13, fontWeight: '600' },
  row: { flexDirection: 'row', gap: 8 },
  sizeChip: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 8,
    borderWidth: 1.5,
    borderColor: Colors.border,
    backgroundColor: Colors.surface,
    minWidth: 44,
    alignItems: 'center',
  },
  sizeChipSelected: { borderColor: Colors.primary, backgroundColor: Colors.primary + '22' },
  sizeChipDisabled: { opacity: 0.35 },
  sizeText: { color: Colors.text, fontWeight: '600', fontSize: 13 },
  sizeTextSelected: { color: Colors.primary },
  sizeTextDisabled: { textDecorationLine: 'line-through' },
  colorChip: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 999,
    borderWidth: 1.5,
    borderColor: Colors.border,
    backgroundColor: Colors.surface,
  },
  colorChipSelected: { borderColor: Colors.accent, backgroundColor: Colors.accent + '22' },
  colorText: { color: Colors.text, fontWeight: '600', fontSize: 13 },
  colorTextSelected: { color: Colors.primary },
  priceRow: { flexDirection: 'row', alignItems: 'center', gap: 12, marginTop: 8 },
  selectedPrice: { color: Colors.primary, fontSize: 22, fontWeight: '800' },
  stockInfo: { color: Colors.success, fontSize: 13, fontWeight: '600' },
});
