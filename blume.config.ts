import { defineConfig } from "blume";

export default defineConfig({
  title: "bsr",
  description:
    "Boy Scout Rule for your linter — suppress baseline violations and report only new ones.",
  logo: {
    text: "bsr",
  },
  github: {
    owner: "sun-yryr",
    repo: "boy-scout-rule-based-lint",
  },
  content: {
    root: "docs",
    sources: [
      { type: "filesystem", root: "docs" },
      {
        type: "github-releases",
        prefix: "changelog",
        owner: "sun-yryr",
        repo: "boy-scout-rule-based-lint",
      },
    ],
  },
  theme: {
    accent: "teal",
    radius: "md",
    mode: "system",
  },
  ai: {
    llmsTxt: true,
  },
  seo: {
    og: { enabled: true },
    rss: { enabled: true, types: ["changelog"] },
    sitemap: true,
    robots: true,
    structuredData: true,
  },
  lastModified: true,
  deployment: {
    output: "static",
    site: "https://bsr.sun-yryr.com",
  },
});
