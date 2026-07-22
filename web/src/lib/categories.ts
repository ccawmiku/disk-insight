export const categoryOrder = [
  "video",
  "audio",
  "image",
  "document",
  "archive",
  "code",
  "data",
  "executable",
  "disk-image",
  "font",
  "design-3d",
  "other",
] as const;

export const categoryColors: Record<string, string> = {
  video: "#ff5a5f",
  audio: "#ff8a00",
  image: "#ffc145",
  document: "#e848a0",
  archive: "#7c3aed",
  code: "#00a8e8",
  data: "#00b894",
  executable: "#ef476f",
  "disk-image": "#118ab2",
  font: "#f78c6b",
  "design-3d": "#8f6ed5",
  other: "#94a3b8",
};
