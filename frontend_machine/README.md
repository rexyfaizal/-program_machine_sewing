# Vue Machine Modular Frontend

Copy folder `src` dan `vite.config.js` ini ke project `C:\rexy\frontend_machine`.

Struktur ini memecah App.vue menjadi:
- api
- composables
- components/dashboard
- components/process
- pages
- utils

Jalankan:

```bash
npm run dev
```

Build untuk deploy ke backend Go:

```bash
npm run build
```

Lalu copy isi folder `dist` ke `backend_machine/public`.
