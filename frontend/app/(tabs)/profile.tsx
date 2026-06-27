import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  StyleSheet,
  Alert,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Colors } from '@/constants/colors';
import { Button } from '@/components/Button';
import { useAuth } from '@/hooks/useAuth';

export default function ProfileScreen() {
  const { user, loading, isVendor, sendOTP, login, logout } = useAuth();
  const router = useRouter();

  const [identifier, setIdentifier] = useState('');
  const [code, setCode] = useState('');
  const [otpSent, setOtpSent] = useState(false);
  const [otpMessage, setOtpMessage] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSendOTP = async () => {
    if (!identifier.trim()) {
      Alert.alert('Error', 'Enter your email or phone number');
      return;
    }
    setSubmitting(true);
    try {
      const res = await sendOTP(identifier.trim());
      setOtpSent(true);
      // Show OTP in dev - remove in production
      setOtpMessage(`Dev OTP: ${res.otp}`);
    } catch (e: any) {
      Alert.alert('Error', e?.response?.data?.error || 'Failed to send OTP');
    } finally {
      setSubmitting(false);
    }
  };

  const handleVerify = async () => {
    if (!code.trim()) {
      Alert.alert('Error', 'Enter the OTP code');
      return;
    }
    setSubmitting(true);
    try {
      await login(identifier.trim(), code.trim());
    } catch (e: any) {
      Alert.alert('Error', e?.response?.data?.error || 'Invalid OTP');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return null;

  if (user) {
    return (
      <ScrollView style={styles.container} contentContainerStyle={styles.content}>
        <View style={styles.card}>
          <Text style={styles.label}>Signed in as</Text>
          <Text style={styles.identifier}>{user.identifier}</Text>
          <View style={[styles.badge, { backgroundColor: isVendor ? Colors.accent + '33' : Colors.primary + '22' }]}>
            <Text style={[styles.badgeText, { color: isVendor ? Colors.primary : Colors.primary }]}>
              {user.role.toUpperCase()}
            </Text>
          </View>
        </View>

        {isVendor && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Vendor Dashboard</Text>
            <Button
              title="Manage Products"
              onPress={() => router.push('/vendor/products')}
              style={styles.mb}
            />
            <Button
              title="View All Orders"
              onPress={() => router.push('/vendor/orders')}
              variant="outline"
            />
          </View>
        )}

        <Button
          title="Sign Out"
          onPress={logout}
          variant="danger"
          style={styles.mt}
        />
      </ScrollView>
    );
  }

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView contentContainerStyle={styles.content}>
        <Text style={styles.title}>Sign In</Text>
        <Text style={styles.sub}>Enter your email or phone number to receive a one-time code.</Text>

        <View style={styles.card}>
          <Text style={styles.inputLabel}>Email or Phone</Text>
          <TextInput
            style={styles.input}
            value={identifier}
            onChangeText={setIdentifier}
            placeholder="you@example.com or +91..."
            placeholderTextColor={Colors.muted}
            autoCapitalize="none"
            keyboardType="email-address"
          />
          <Button
            title="Send OTP"
            onPress={handleSendOTP}
            loading={submitting && !otpSent}
            style={styles.mt}
          />
          {otpMessage ? <Text style={styles.otpMsg}>{otpMessage}</Text> : null}
        </View>

        {otpSent && (
          <View style={styles.card}>
            <Text style={styles.inputLabel}>Enter OTP Code</Text>
            <TextInput
              style={styles.input}
              value={code}
              onChangeText={setCode}
              placeholder="6-digit code"
              placeholderTextColor={Colors.muted}
              keyboardType="number-pad"
              maxLength={6}
            />
            <Button
              title="Verify & Sign In"
              onPress={handleVerify}
              loading={submitting}
              style={styles.mt}
            />
          </View>
        )}
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: Colors.background },
  content: { padding: 20, gap: 16 },
  title: { color: Colors.text, fontSize: 26, fontWeight: '800' },
  sub: { color: Colors.muted, fontSize: 14 },
  card: {
    backgroundColor: Colors.surface,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: Colors.border,
    padding: 20,
    gap: 10,
  },
  label: { color: Colors.muted, fontSize: 13 },
  identifier: { color: Colors.text, fontSize: 18, fontWeight: '700' },
  badge: { alignSelf: 'flex-start', paddingHorizontal: 12, paddingVertical: 4, borderRadius: 999, marginTop: 4 },
  badgeText: { fontSize: 12, fontWeight: '700' },
  section: { gap: 12 },
  sectionTitle: { color: Colors.text, fontSize: 18, fontWeight: '700' },
  inputLabel: { color: Colors.muted, fontSize: 13, fontWeight: '600' },
  input: {
    backgroundColor: Colors.background,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 10,
    padding: 12,
    color: Colors.text,
    fontSize: 15,
  },
  otpMsg: { color: Colors.success, fontSize: 13, fontWeight: '600' },
  mt: { marginTop: 8 },
  mb: { marginBottom: 4 },
});
