import React, { useState, useRef } from 'react';
import { View, Image, ScrollView, StyleSheet, TouchableOpacity, useWindowDimensions, NativeSyntheticEvent, NativeScrollEvent } from 'react-native';
import { Colors } from '@/constants/colors';

interface Props {
  images: string[];
  fallback?: string;
}

export function ImageGallery({ images, fallback }: Props) {
  const { width } = useWindowDimensions();
  const imgWidth = Math.min(width, 600);
  const [activeIndex, setActiveIndex] = useState(0);
  const scrollRef = useRef<ScrollView>(null);

  const allImages = images.length > 0 ? images : fallback ? [fallback] : [];

  const handleScroll = (e: NativeSyntheticEvent<NativeScrollEvent>) => {
    const index = Math.round(e.nativeEvent.contentOffset.x / imgWidth);
    setActiveIndex(index);
  };

  if (allImages.length === 0) {
    return (
      <View style={[styles.placeholder, { width: imgWidth, height: imgWidth * 0.75 }]}>
      </View>
    );
  }

  return (
    <View>
      <ScrollView
        ref={scrollRef}
        horizontal
        pagingEnabled
        showsHorizontalScrollIndicator={false}
        onScroll={handleScroll}
        scrollEventThrottle={16}
      >
        {allImages.map((url, i) => (
          <Image
            key={i}
            source={{ uri: url }}
            style={{ width: imgWidth, height: imgWidth * 0.75 }}
            resizeMode="cover"
          />
        ))}
      </ScrollView>

      {allImages.length > 1 && (
        <View style={styles.dots}>
          {allImages.map((_, i) => (
            <TouchableOpacity
              key={i}
              style={[styles.dot, i === activeIndex && styles.dotActive]}
              onPress={() => {
                scrollRef.current?.scrollTo({ x: i * imgWidth, animated: true });
                setActiveIndex(i);
              }}
            />
          ))}
        </View>
      )}

      {/* Thumbnail strip */}
      {allImages.length > 1 && (
        <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.thumbnails}>
          {allImages.map((url, i) => (
            <TouchableOpacity
              key={i}
              onPress={() => {
                scrollRef.current?.scrollTo({ x: i * imgWidth, animated: true });
                setActiveIndex(i);
              }}
            >
              <Image
                source={{ uri: url }}
                style={[styles.thumb, i === activeIndex && styles.thumbActive]}
                resizeMode="cover"
              />
            </TouchableOpacity>
          ))}
        </ScrollView>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  placeholder: {
    backgroundColor: Colors.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  dots: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 6,
    paddingVertical: 8,
  },
  dot: {
    width: 7,
    height: 7,
    borderRadius: 999,
    backgroundColor: Colors.border,
  },
  dotActive: {
    backgroundColor: Colors.primary,
    width: 18,
  },
  thumbnails: {
    paddingHorizontal: 12,
    paddingVertical: 8,
    gap: 8,
  },
  thumb: {
    width: 60,
    height: 60,
    borderRadius: 8,
    borderWidth: 1.5,
    borderColor: Colors.border,
  },
  thumbActive: {
    borderColor: Colors.primary,
  },
});
