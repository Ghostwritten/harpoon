# Harpoon Website

Documentation site built with [VitePress](https://vitepress.dev/).

## Local development

```bash
npm install
npm run docs:dev
```

## Build

```bash
npm run docs:build
```

## Deployment

The site is deployed automatically via GitHub Actions when changes are pushed to `main` in the `website/` directory. The workflow builds the site and deploys to the `gh-pages` branch.

To enable GitHub Pages:

1. Go to **Settings → Pages** in the repository
2. Set **Source** to "Deploy from a branch"
3. Set **Branch** to `gh-pages` and root directory to `/`
4. Save

The site will be available at `https://ghostwritten.github.io/harpoon/`.
