import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TextInput,
  Alert,
  Modal,
  TouchableOpacity,
  ScrollView,
  ActivityIndicator,
  useWindowDimensions,
} from 'react-native';
import { Colors } from '@/constants/colors';
import { Button } from '@/components/Button';
import {
  vendorListProducts,
  vendorCreateProduct,
  vendorUpdateProduct,
  vendorDeactivateProduct,
} from '@/services/api';
import { Ionicons } from '@expo/vector-icons';

interface Product {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  image_url: string;
  stock: number;
  is_active: boolean;
}

const emptyForm = { name: '', description: '', price: '', image_url: '', stock: '' };

export default function VendorProductsScreen() {
  const { width } = useWindowDimensions();
  const isWide = width >= 768;
  const numColumns = isWide ? 2 : 1;

  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [editProduct, setEditProduct] = useState<Product | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [saving, setSaving] = useState(false);

  const fetch = useCallback(async () => {
    try {
      const res = await vendorListProducts();
      setProducts(res.data || []);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);

  const openCreate = () => {
    setEditProduct(null);
    setForm(emptyForm);
    setModalVisible(true);
  };

  const openEdit = (p: Product) => {
    setEditProduct(p);
    setForm({
      name: p.name,
      description: p.description,
      price: String(p.price_cents / 100),
      image_url: p.image_url,
      stock: String(p.stock),
    });
    setModalVisible(true);
  };

  const handleSave = async () => {
    if (!form.name.trim() || !form.price) {
      Alert.alert('Error', 'Name and price are required.');
      return;
    }
    const priceCents = Math.round(parseFloat(form.price) * 100);
    if (isNaN(priceCents) || priceCents <= 0) {
      Alert.alert('Error', 'Enter a valid price.');
      return;
    }

    setSaving(true);
    try {
      if (editProduct) {
        await vendorUpdateProduct(editProduct.id, {
          name: form.name,
          description: form.description,
          price_cents: priceCents,
          image_url: form.image_url,
          stock: parseInt(form.stock) || 0,
        });
      } else {
        await vendorCreateProduct({
          name: form.name,
          description: form.description,
          price_cents: priceCents,
          image_url: form.image_url,
          stock: parseInt(form.stock) || 0,
        });
      }
      setModalVisible(false);
      fetch();
    } catch (e: any) {
      Alert.alert('Error', e?.response?.data?.error || 'Failed to save product');
    } finally {
      setSaving(false);
    }
  };

  const handleDeactivate = (p: Product) => {
    Alert.alert('Deactivate Product', `Remove "${p.name}" from the store?`, [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Deactivate',
        style: 'destructive',
        onPress: async () => {
          await vendorDeactivateProduct(p.id);
          fetch();
        },
      },
    ]);
  };

  if (loading) {
    return <View style={styles.center}><ActivityIndicator color={Colors.primary} size="large" /></View>;
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Products ({products.length})</Text>
        <Button title="+ Add Product" onPress={openCreate} style={styles.addBtn} />
      </View>

      <FlatList
        data={products}
        key={numColumns}
        keyExtractor={(p) => String(p.id)}
        numColumns={numColumns}
        columnWrapperStyle={numColumns > 1 ? styles.row : undefined}
        contentContainerStyle={styles.list}
        renderItem={({ item }) => (
          <View style={[styles.card, !item.is_active && styles.inactive, numColumns > 1 && styles.cardWide]}>
            <View style={styles.cardHeader}>
              <Text style={styles.productName} numberOfLines={1}>{item.name}</Text>
              {!item.is_active && <Text style={styles.inactiveBadge}>INACTIVE</Text>}
            </View>
            <Text style={styles.price}>${(item.price_cents / 100).toFixed(2)}</Text>
            <Text style={styles.meta}>Stock: {item.stock}</Text>
            <Text style={styles.desc} numberOfLines={2}>{item.description}</Text>
            <View style={styles.actions}>
              <Button title="Edit" onPress={() => openEdit(item)} variant="outline" style={styles.actionBtn} />
              {item.is_active && (
                <Button title="Deactivate" onPress={() => handleDeactivate(item)} variant="danger" style={styles.actionBtn} />
              )}
            </View>
          </View>
        )}
        ListEmptyComponent={<Text style={styles.empty}>No products yet. Add your first product.</Text>}
      />

      <Modal visible={modalVisible} animationType="slide" presentationStyle="pageSheet">
        <ScrollView style={styles.modal} contentContainerStyle={styles.modalContent}>
          <View style={styles.modalHeader}>
            <Text style={styles.modalTitle}>{editProduct ? 'Edit Product' : 'New Product'}</Text>
            <TouchableOpacity onPress={() => setModalVisible(false)}>
              <Ionicons name="close" size={24} color={Colors.text} />
            </TouchableOpacity>
          </View>

          {[
            { label: 'Name *', key: 'name', placeholder: 'Product name' },
            { label: 'Description', key: 'description', placeholder: 'Product description', multiline: true },
            { label: 'Price ($) *', key: 'price', placeholder: '29.99', keyboardType: 'decimal-pad' },
            { label: 'Stock Quantity', key: 'stock', placeholder: '100', keyboardType: 'number-pad' },
            { label: 'Image URL', key: 'image_url', placeholder: 'https://...' },
          ].map((field) => (
            <View key={field.key} style={styles.fieldGroup}>
              <Text style={styles.fieldLabel}>{field.label}</Text>
              <TextInput
                style={[styles.input, field.multiline && styles.textarea]}
                value={(form as any)[field.key]}
                onChangeText={(v) => setForm((f) => ({ ...f, [field.key]: v }))}
                placeholder={field.placeholder}
                placeholderTextColor={Colors.muted}
                multiline={field.multiline}
                keyboardType={(field as any).keyboardType || 'default'}
                autoCapitalize="none"
              />
            </View>
          ))}

          <Button title="Save Product" onPress={handleSave} loading={saving} style={styles.mt} />
          <Button title="Cancel" onPress={() => setModalVisible(false)} variant="outline" style={styles.mt} />
        </ScrollView>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: Colors.background },
  header: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', padding: 20, paddingBottom: 8 },
  title: { color: Colors.text, fontSize: 20, fontWeight: '800' },
  addBtn: { paddingVertical: 8, paddingHorizontal: 14, minHeight: 36 },
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
  inactive: { opacity: 0.5 },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  productName: { color: Colors.text, fontSize: 16, fontWeight: '700', flex: 1 },
  inactiveBadge: { fontSize: 11, color: Colors.error, fontWeight: '700', backgroundColor: Colors.error + '22', paddingHorizontal: 8, paddingVertical: 2, borderRadius: 999 },
  price: { color: Colors.primary, fontSize: 18, fontWeight: '800' },
  meta: { color: Colors.muted, fontSize: 13 },
  desc: { color: Colors.muted, fontSize: 13 },
  actions: { flexDirection: 'row', gap: 8, marginTop: 8 },
  actionBtn: { flex: 1, paddingVertical: 8, minHeight: 36 },
  empty: { color: Colors.muted, textAlign: 'center', marginTop: 48, fontSize: 16 },
  modal: { flex: 1, backgroundColor: Colors.background },
  modalContent: { padding: 24, gap: 16 },
  modalHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 },
  modalTitle: { color: Colors.text, fontSize: 22, fontWeight: '800' },
  fieldGroup: { gap: 6 },
  fieldLabel: { color: Colors.muted, fontSize: 13, fontWeight: '600' },
  input: {
    backgroundColor: Colors.surface,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 10,
    padding: 12,
    color: Colors.text,
    fontSize: 15,
  },
  textarea: { height: 80, textAlignVertical: 'top' },
  mt: { marginTop: 4 },
});
