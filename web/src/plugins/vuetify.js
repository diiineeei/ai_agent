import { createVuetify } from 'vuetify'
import 'vuetify/styles'

export default createVuetify({
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary:            '#1565C0',
          'primary-darken-1': '#0D47A1',
          secondary:          '#1E88E5',
          background:         '#F8F8FC',
          surface:            '#FFFFFF',
          'surface-variant':  '#EEEEF6',
          'on-surface-variant': '#3D3B6B',
          error:              '#D32F2F',
          info:               '#5C54E0',
          success:            '#2E7D32',
          warning:            '#E65100',
        },
      },
      dark: {
        colors: {
          primary:   '#6C63FF',
          secondary: '#9C94FF',
          background: '#0F0E1A',
          surface:    '#1A1929',
          'surface-variant': '#252438',
        },
      },
    },
  },
})
