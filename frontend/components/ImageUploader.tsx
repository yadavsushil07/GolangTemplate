"use client";

import { useCallback, useRef, useState } from "react";
import { uploadToCloudinary, vendorDeleteImage } from "@/lib/api";

interface ExistingImage {
  id: number;
  url: string;
  sort_order?: number;
}

interface UploadingFile {
  id: string; // temporary local id
  name: string;
  progress: number; // 0-100
  error?: string;
  url?: string; // set after success
}

interface Props {
  /** If provided the component is in "live edit" mode and deletes/adds images immediately via the API */
  productId?: number;
  /** Pre-existing images from the DB (edit mode) */
  existingImages?: ExistingImage[];
  /** Called in create mode whenever the ready-to-save URL list changes */
  onImagesChange?: (urls: string[]) => void;
  maxImages?: number;
}

const CLOUD_NAME = process.env.NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME ?? "";
const UPLOAD_PRESET = process.env.NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET ?? "";

export default function ImageUploader({
  productId,
  existingImages = [],
  onImagesChange,
  maxImages = 8,
}: Props) {
  const [saved, setSaved] = useState<ExistingImage[]>(existingImages);
  const [uploading, setUploading] = useState<UploadingFile[]>([]);
  const [pendingUrls, setPendingUrls] = useState<string[]>([]);
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const notifyChange = useCallback(
    (urls: string[]) => {
      onImagesChange?.(urls);
    },
    [onImagesChange]
  );

  async function processFiles(files: FileList | File[]) {
    const arr = Array.from(files).filter((f) => f.type.startsWith("image/"));
    if (!arr.length) return;

    if (!CLOUD_NAME || !UPLOAD_PRESET) {
      alert(
        "Cloudinary is not configured.\n\nAdd NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME and NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET to your frontend/.env.local file."
      );
      return;
    }

    const slots: UploadingFile[] = arr.map((f) => ({
      id: Math.random().toString(36).slice(2),
      name: f.name,
      progress: 0,
    }));
    setUploading((prev) => [...prev, ...slots]);

    for (let i = 0; i < arr.length; i++) {
      const file = arr[i];
      const slot = slots[i];
      try {
        const url = await uploadToCloudinary(
          file,
          CLOUD_NAME,
          UPLOAD_PRESET,
          (pct) => {
            setUploading((prev) =>
              prev.map((u) => (u.id === slot.id ? { ...u, progress: pct } : u))
            );
          }
        );

        setUploading((prev) =>
          prev.map((u) => (u.id === slot.id ? { ...u, progress: 100, url } : u))
        );

        if (productId) {
          // live mode — attach to product immediately via API
          const { vendorAddImages } = await import("@/lib/api");
          await vendorAddImages(productId, [url]);
          setSaved((prev) => [...prev, { id: Date.now(), url }]);
          setUploading((prev) => prev.filter((u) => u.id !== slot.id));
        } else {
          // create mode — buffer URLs and notify parent
          setPendingUrls((prev) => {
            const next = [...prev, url];
            notifyChange(next);
            return next;
          });
          setUploading((prev) => prev.filter((u) => u.id !== slot.id));
        }
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Upload failed";
        setUploading((prev) =>
          prev.map((u) => (u.id === slot.id ? { ...u, error: msg } : u))
        );
      }
    }
  }

  async function removeSaved(img: ExistingImage) {
    if (productId) {
      try {
        await vendorDeleteImage(productId, img.id);
      } catch {
        alert("Failed to delete image");
        return;
      }
    }
    setSaved((prev) => prev.filter((i) => i.id !== img.id));
  }

  function removePending(url: string) {
    setPendingUrls((prev) => {
      const next = prev.filter((u) => u !== url);
      notifyChange(next);
      return next;
    });
  }

  function clearError(id: string) {
    setUploading((prev) => prev.filter((u) => u.id !== id));
  }

  const totalCount = saved.length + pendingUrls.length + uploading.filter((u) => !u.error).length;
  const canAdd = totalCount < maxImages;

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragging(false);
      if (canAdd) processFiles(e.dataTransfer.files);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [canAdd]
  );

  return (
    <div className="space-y-3">
      {/* Thumbnail grid */}
      {(saved.length > 0 || pendingUrls.length > 0 || uploading.length > 0) && (
        <div className="flex flex-wrap gap-2">
          {/* Saved (from DB or just uploaded in live mode) */}
          {saved.map((img) => (
            <div key={img.id} className="relative group w-20 h-20 flex-shrink-0">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={img.url}
                alt=""
                className="w-full h-full object-cover bg-[#ECE3D5] rounded-sm"
              />
              <button
                type="button"
                onClick={() => removeSaved(img)}
                className="absolute top-0.5 right-0.5 w-5 h-5 rounded-full bg-black/60 text-white text-[10px] flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                title="Remove"
              >
                ×
              </button>
            </div>
          ))}

          {/* Pending (create mode — not yet saved to DB) */}
          {pendingUrls.map((url) => (
            <div key={url} className="relative group w-20 h-20 flex-shrink-0">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={url}
                alt=""
                className="w-full h-full object-cover bg-[#ECE3D5] rounded-sm"
              />
              <span className="absolute bottom-0 left-0 right-0 text-center text-[8px] bg-black/40 text-white py-0.5">
                queued
              </span>
              <button
                type="button"
                onClick={() => removePending(url)}
                className="absolute top-0.5 right-0.5 w-5 h-5 rounded-full bg-black/60 text-white text-[10px] flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                title="Remove"
              >
                ×
              </button>
            </div>
          ))}

          {/* In-progress uploads */}
          {uploading.map((u) => (
            <div key={u.id} className="relative w-20 h-20 flex-shrink-0 bg-[#ECE3D5] rounded-sm flex flex-col items-center justify-center p-1">
              {u.error ? (
                <>
                  <span className="text-red-500 text-[9px] text-center leading-tight">{u.error}</span>
                  <button
                    type="button"
                    onClick={() => clearError(u.id)}
                    className="text-[9px] text-[#8B8175] underline mt-1"
                  >
                    dismiss
                  </button>
                </>
              ) : (
                <>
                  <div className="w-12 h-1.5 bg-white/60 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-[#B06A50] transition-all"
                      style={{ width: `${u.progress}%` }}
                    />
                  </div>
                  <span className="text-[9px] text-[#43293A] mt-1">{u.progress}%</span>
                  <span className="text-[8px] text-[#8B8175] truncate w-full text-center mt-0.5">{u.name}</span>
                </>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Drop zone */}
      {canAdd && (
        <div
          onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
          onDragLeave={() => setDragging(false)}
          onDrop={onDrop}
          onClick={() => inputRef.current?.click()}
          className={`border-2 border-dashed rounded-sm p-5 text-center cursor-pointer transition-colors ${
            dragging
              ? "border-[#B06A50] bg-[#fdf8f5]"
              : "border-[#E4DAC9] hover:border-[#B06A50] hover:bg-[#fdf8f5]"
          }`}
        >
          <input
            ref={inputRef}
            type="file"
            accept="image/*"
            multiple
            className="sr-only"
            onChange={(e) => e.target.files && processFiles(e.target.files)}
          />
          <p className="text-sm text-[#8B8175]">
            Drag & drop images here, or{" "}
            <span className="text-[#B06A50] underline">browse</span>
          </p>
          <p className="text-[10px] text-[#8B8175] mt-1">
            JPG, PNG, WEBP · max {maxImages} images
          </p>
        </div>
      )}
    </div>
  );
}
