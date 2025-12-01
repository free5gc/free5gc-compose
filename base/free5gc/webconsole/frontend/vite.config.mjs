import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";

export default defineConfig({
  base: "/",
  plugins: [react({})],

  build: {
    outDir: "build",
  },
});
