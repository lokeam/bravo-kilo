import { Controller } from 'react-hook-form';

type LangaugeOption = {
  value: string,
  label: string,
};

const languages: LangaugeOption[] = [
  { value: 'en', label: 'English'  },
  { value: 'zh', label: '中文' },
  { value: 'hi', label: 'हिन्दी'  },
  { value: 'es', label: 'Español' },
  { value: 'ca', label: 'Català' },
  { value: 'fr', label: 'Français'  },
  { value: 'ar', label: 'العربية' },
  { value: 'bn', label: 'বাংলা' },
  { value: 'pt', label: 'Português' },
  { value: 'ru', label: 'Русский'  },
  { value: 'ur', label: 'اردو' },
  { value: 'id', label: 'Bahasa Indonesia'  },
  { value: 'de', label: 'Deutsch' },
  { value: 'ja', label: '日本語'  },
  { value: 'sw', label: 'Kiswahili' },
  { value: 'pnb', label: 'شاہ مکھی پنجابی (Shāhmukhī Pañjābī)' },
  { value: 'mr', label: 'मराठी' },
  { value: 'te', label: 'తెలుగు' },
  { value: 'hu', label: 'Magyar' },
  { value: 'tr', label: 'Türkçe'  },
  { value: 'ko', label: '한국어' },
  { value: 'ta', label: 'தமிழ்' },
  { value: 'el', label: 'Ελληνικά' },
  { value: 'vi', label: 'Tiếng Việt' },
  { value: 'it', label: 'Italiano'  },
  { value: 'th', label: 'ไทย' },
  { value: 'fa', label: 'فارسی' },
  { value: 'gu', label: 'ગુજરાતી' },
  { value: 'pl', label: 'Polski' },
  { value: 'uk', label: 'Українська' },
  { value: 'ml', label: 'മലയാളം' },
  { value: 'kn', label: 'ಕನ್ನಡ' },
  { value: 'ro', label: 'Română' },
  { value: 'he', label: 'עברית' },
  { value: 'eu', label: 'Euskara' },
  { value: 'my', label: 'မြန်မာဘာသာ' },
  { value: 'xh', label: 'isiXhosa' },
  { value: 'so', label: 'Soomaali' },
  { value: 'ne', label: 'नेपाली' },
  { value: 'no', label: 'Norsk (Bokmål' },
  { value: 'tl', label: 'Tagalog' },
  { value: 'fi', label: 'Suomi' },
  { value: 'rw', label: 'Ikinyarwanda' },
  { value: 'ng', label: 'Oshiwambo' },
  { value: 'sl', label: 'Slovenščina' },
  { value: 'ga', label: 'Gaeilge' },
  { value: 'hy', label: 'Հայերեն' },
  { value: 'da', label: 'Dansk' },
  { value: 'nl', label: 'Nederlands' },
  { value: 'af', label: 'Afrikaans' },
];

function LanguageSelect({ control, errors }) {

  return (
    <div className="block col-span-2">
      <label htmlFor="language" className="block mb-2 text-base font-medium text-gray-900 dark:text-white">
        Language <span className="text-red-600 ml-px">*</span>
      </label>
      <Controller
        name="language"
        control={control}
        render={({ field }) => (
          <select
            {...field}
            className={`bg-maastricht border ${errors.language ? 'border-red-500' : 'border-gray-600'} text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:text-white`}
          >
            {languages.map((language) => (
              <option key={language.value} value={language.value}>
                {language.label}
              </option>
            ))}
          </select>
        )}
      />
      {errors.language && <p className="text-red-500">{errors.language.message}</p>}
    </div>
  );
}


export default LanguageSelect;
