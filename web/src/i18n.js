import { reactive, computed } from 'vue'

const state = reactive({
  lang: localStorage.getItem('lang') || 'en'
})

export const useI18n = () => {
  const setLang = (l) => {
    state.lang = l
    localStorage.setItem('lang', l)
  }

  const t = (key) => {
    const keys = key.split('.')
    let val = translations[state.lang]
    for (const k of keys) {
      val = val[k]
    }
    return val
  }

  return {
    lang: computed(() => state.lang),
    setLang,
    t
  }
}

const translations = {
  en: {
    nav: {
      home: 'Home',
      docs: 'Docs',
      github: 'GitHub'
    },
    hero: {
      title: 'AI Coding Agent for your Terminal',
      subtitle: 'CuRe Code is an autonomous command-line agent that reads, writes, and refactors code. Built with Go for performance and safety.',
      start: 'Get Started',
      install: 'Install CLI'
    },
    features: {
      loop: {
        title: 'Autonomous Loop',
        desc: 'Autonomously reads files, runs commands, and applies edits until the task is complete.'
      },
      edits: {
        title: 'Precision Edits',
        desc: 'Surgical search-and-replace that respects your project indentation and linting rules.'
      },
      context: {
        title: 'Context Aware',
        desc: 'Mention files with @ or let the agent discover them using high-performance globbing.'
      },
      safe: {
        title: 'Safe by Design',
        desc: 'Every shell command and destructive edit requires explicit user confirmation.'
      },
      plug: {
        title: 'Plug & Play',
        desc: 'Supports Gemini, OpenAI, Claude, and local models via Ollama out of the box.'
      },
      binary: {
        title: 'Static Binary',
        desc: 'Zero dependencies. Just a single Go executable that runs anywhere.'
      }
    },
    install: {
      title: 'Quick Install',
      subtitle: 'Run this command to download and install CuRe Code automatically:',
      win: 'Download the binary and run the installer in your terminal:',
      win_link: 'Download curecode.exe from Releases'
    },
    contributors: {
      title: 'Contributors',
      subtitle: 'CuRe Code is an open-source project. We welcome all contributions!',
      cta: 'Join the development and help us make the best AI coding agent.'
    }
  },
  id: {
    nav: {
      home: 'Beranda',
      docs: 'Dokumentasi',
      github: 'GitHub'
    },
    hero: {
      title: 'Agen AI Coding untuk Terminal Anda',
      subtitle: 'CuRe Code adalah agen baris perintah otonom yang membaca, menulis, dan merombak kode. Dibangun dengan Go untuk performa dan keamanan.',
      start: 'Mulai Sekarang',
      install: 'Instal CLI'
    },
    features: {
      loop: {
        title: 'Loop Otonom',
        desc: 'Secara mandiri membaca file, menjalankan perintah, dan menerapkan pengeditan hingga tugas selesai.'
      },
      edits: {
        title: 'Edit Presisi',
        desc: 'Pencarian-dan-penggantian bedah yang menghormati indentasi proyek dan aturan linting Anda.'
      },
      context: {
        title: 'Sadar Konteks',
        desc: 'Sebutkan file dengan @ atau biarkan agen menemukannya menggunakan globbing performa tinggi.'
      },
      safe: {
        title: 'Aman secara Desain',
        desc: 'Setiap perintah shell dan pengeditan destruktif memerlukan konfirmasi pengguna secara eksplisit.'
      },
      plug: {
        title: 'Plug & Play',
        desc: 'Mendukung Gemini, OpenAI, Claude, dan model lokal melalui Ollama secara langsung.'
      },
      binary: {
        title: 'Binary Statis',
        desc: 'Tanpa dependensi. Cukup satu executable Go yang berjalan di mana saja.'
      }
    },
    install: {
      title: 'Instalasi Cepat',
      subtitle: 'Jalankan perintah ini untuk mengunduh dan menginstal CuRe Code secara otomatis:',
      win: 'Unduh binary dan jalankan installer di terminal Anda:',
      win_link: 'Unduh curecode.exe dari Releases'
    },
    contributors: {
      title: 'Kontributor',
      subtitle: 'CuRe Code adalah proyek sumber terbuka. Kami menyambut semua kontribusi!',
      cta: 'Bergabunglah dalam pengembangan dan bantu kami membuat agen coding AI terbaik.'
    }
  }
}
