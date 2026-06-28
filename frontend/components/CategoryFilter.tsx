import React from 'react';
import { ScrollView, TouchableOpacity, Text, StyleSheet, View } from 'react-native';
import { Colors } from '@/constants/colors';

interface Category {
  id: number;
  name: string;
  slug: string;
}

interface Props {
  categories: Category[];
  selected: string;
  onSelect: (slug: string) => void;
}

export function CategoryFilter({ categories, selected, onSelect }: Props) {
  return (
    <View style={styles.wrapper}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.container}>
        <TouchableOpacity
          style={[styles.chip, selected === '' && styles.chipActive]}
          onPress={() => onSelect('')}
        >
          <Text style={[styles.chipText, selected === '' && styles.chipTextActive]}>All</Text>
        </TouchableOpacity>
        {categories.map((cat) => (
          <TouchableOpacity
            key={cat.id}
            style={[styles.chip, selected === cat.slug && styles.chipActive]}
            onPress={() => onSelect(cat.slug)}
          >
            <Text style={[styles.chipText, selected === cat.slug && styles.chipTextActive]}>
              {cat.name}
            </Text>
          </TouchableOpacity>
        ))}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    borderBottomWidth: 1,
    borderColor: Colors.border,
    backgroundColor: Colors.surface,
  },
  container: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    gap: 8,
    flexDirection: 'row',
  },
  chip: {
    paddingHorizontal: 14,
    paddingVertical: 7,
    borderRadius: 999,
    borderWidth: 1.5,
    borderColor: Colors.border,
    backgroundColor: 'transparent',
  },
  chipActive: {
    borderColor: Colors.primary,
    backgroundColor: Colors.primary + '22',
  },
  chipText: {
    color: Colors.muted,
    fontSize: 13,
    fontWeight: '600',
  },
  chipTextActive: {
    color: Colors.primary,
  },
});
