import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  StyleSheet,
  ActivityIndicator,
  TouchableOpacity,
  Linking,
  TextInput,
  Alert,
} from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { getProduct } from '@/services/api';
import { useCart } from '@/hooks/useCart';
import { ImageGallery } from '@/components/ImageGallery';
import { VariantSelector } from '@/components/VariantSelector';
import { Button } from '@/components/Button';

const WHATSAPP_NUMBER = '919999999999'; // replace with vendor number

export default function ProductDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { add } = useCart();

  const [product, setProduct] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [selectedVariantId, setSelectedVariantId] = useState<number | null>(null);
  const [note, setNote] = useState('');
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const res = await getProduct(Number(id)) as any;
        setProduct(res.data);
        if (res.data?.variants?.length > 0) {
          const first = res.data.variants.find((v: any) => v.is_active && v.stock > 0);
          if (first) setSelectedVariantId(first.id);
        }
      } catch {
        // no-op
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  if (loading) {
    return <View style={styles.center}><ActivityIndicator color={Colors.primary} size="large" /></View>;
  }
  if (!product) {
    return <View style={styles.center}><Text style={styles.errorText}>Product not found.</Text></View>;
  }

  const imageUrls = (product.images ?? []).map((img: any) => img.url);
  const hasVariants = product.variants && product.variants.length > 0;
  const selectedVariant = hasVariants
    ? product.variants.find((v: any) => v.id === selectedVariantId)
    : null;

  const price = selectedVariant
    ? selectedVariant.price_cents
    : product.price_cents;
  const inStock = selectedVariant
    ? selectedVariant.stock > 0
    : product.stock > 0;

  const handleAddToCart = async () => {
    if (hasVariants && !selectedVariantId) {
      Alert.alert('Select variant', 'Please select a size before adding to cart.');
      return;
    }
    setAdding(true);
    try {
      await add(product.id, selectedVariantId ?? undefined);
      router.push('/cart');
    } finally {
      setAdding(false);
    }
  };

  const handleWhatsApp = () => {
    const variantText = selectedVariant
      ? `\nSize: ${selectedVariant.size}${selectedVariant.color ? ` | Color: ${selectedVariant.color}` : ''}`
      : '';
    const noteText = note ? `\nNote: ${note}` : '';
    const msg = encodeURIComponent(
      `Hi! I want to order:\n*${product.name}*${variantText}\nPrice: ₹${(price / 100).toLocaleString('en-IN')}${noteText}`
    );
    Linking.openURL(`https://wa.me/${WHATSAPP_NUMBER}?text=${msg}`);
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <ImageGallery images={imageUrls} fallback={product.image_url} />

      <View style={styles.info}>
        {product.categories?.length > 0 && (
          <Text style={styles.categoryLabel}>
            {product.categories.map((c: any) => c.name).join(' · ')}
          </Text>
        )}
        <Text style={styles.name}>{product.name}</Text>
        {!hasVariants && (
          <Text style={styles.price}>₹{(price / 100).toLocaleString('en-IN')}</Text>
        )}
        <Text style={styles.description}>{product.description}</Text>

        {hasVariants && (
          <View style={styles.section}>
            <VariantSelector
              variants={product.variants}
              selectedVariantId={selectedVariantId}
              onSelect={(v) => setSelectedVariantId(v.id)}
            />
          </View>
        )}

        <View style={styles.section}>
          <Text style={styles.sectionLabel}>Customization Note (optional)</Text>
          <TextInput
            style={styles.noteInput}
            value={note}
            onChangeText={setNote}
            placeholder="E.g. embroidery name, color preference..."
            placeholderTextColor={Colors.muted}
            multiline
            numberOfLines={3}
          />
        </View>

        <View style={styles.actions}>
          <Button
            title={adding ? 'Adding...' : inStock ? 'Add to Cart' : 'Out of Stock'}
            onPress={handleAddToCart}
            disabled={!inStock || adding}
            style={styles.addBtn}
          />
          <TouchableOpacity style={styles.whatsappBtn} onPress={handleWhatsApp}>
            <Text style={styles.whatsappText}>Order via WhatsApp</Text>
          </TouchableOpacity>
        </View>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  content: { paddingBottom: 40 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: Colors.background },
  errorText: { color: Colors.error, fontSize: 16 },
  info: { padding: 20, gap: 12 },
  categoryLabel: { color: Colors.primary, fontSize: 12, fontWeight: '700', textTransform: 'uppercase', letterSpacing: 1 },
  name: { color: Colors.text, fontSize: 24, fontWeight: '800' },
  price: { color: Colors.primary, fontSize: 26, fontWeight: '800' },
  description: { color: Colors.muted, fontSize: 15, lineHeight: 22 },
  section: { marginTop: 8 },
  sectionLabel: { color: Colors.muted, fontSize: 13, fontWeight: '600', marginBottom: 8 },
  noteInput: {
    backgroundColor: Colors.surface,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 10,
    padding: 12,
    color: Colors.text,
    fontSize: 14,
    minHeight: 80,
    textAlignVertical: 'top',
  },
  actions: { gap: 12, marginTop: 8 },
  addBtn: { minHeight: 52 },
  whatsappBtn: {
    backgroundColor: '#25D366',
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 14,
  },
  whatsappText: { color: '#fff', fontWeight: '700', fontSize: 16 },
});
