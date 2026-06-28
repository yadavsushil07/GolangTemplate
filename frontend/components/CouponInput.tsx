import React, { useState } from 'react';
import { View, TextInput, StyleSheet, Text, ActivityIndicator } from 'react-native';
import { Colors } from '@/constants/colors';
import { Button } from './Button';
import { validateCoupon } from '@/services/api';

interface Props {
  orderTotalCents: number;
  onDiscountApplied: (discountCents: number, code: string) => void;
}

export function CouponInput({ orderTotalCents, onDiscountApplied }: Props) {
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [isValid, setIsValid] = useState(false);

  const handleApply = async () => {
    if (!code.trim()) return;
    setLoading(true);
    setMessage('');
    try {
      const res = await validateCoupon(code.trim().toUpperCase(), orderTotalCents);
      const data = res.data;
      setIsValid(data.valid);
      setMessage(data.message);
      if (data.valid) {
        onDiscountApplied(data.discount_cents, code.trim().toUpperCase());
      } else {
        onDiscountApplied(0, '');
      }
    } catch {
      setMessage('Failed to validate coupon');
      setIsValid(false);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.label}>Discount Coupon</Text>
      <View style={styles.row}>
        <TextInput
          style={styles.input}
          value={code}
          onChangeText={(t) => {
            setCode(t.toUpperCase());
            setMessage('');
            setIsValid(false);
            onDiscountApplied(0, '');
          }}
          placeholder="Enter coupon code"
          placeholderTextColor={Colors.muted}
          autoCapitalize="characters"
        />
        {loading ? (
          <ActivityIndicator color={Colors.primary} style={styles.loader} />
        ) : (
          <Button title="Apply" onPress={handleApply} style={styles.btn} />
        )}
      </View>
      {message ? (
        <Text style={[styles.msg, isValid ? styles.msgSuccess : styles.msgError]}>{message}</Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 8 },
  label: { color: Colors.muted, fontSize: 13, fontWeight: '600' },
  row: { flexDirection: 'row', gap: 8, alignItems: 'center' },
  input: {
    flex: 1,
    backgroundColor: Colors.background,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 10,
    padding: 12,
    color: Colors.text,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 1,
  },
  btn: { paddingVertical: 10, paddingHorizontal: 16, minHeight: 44 },
  loader: { paddingHorizontal: 16 },
  msg: { fontSize: 13, fontWeight: '600' },
  msgSuccess: { color: Colors.success },
  msgError: { color: Colors.error },
});
