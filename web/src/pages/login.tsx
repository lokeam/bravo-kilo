import { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { ErrorType, ERROR_MESSAGES, DEFAULT_ERROR_MESSAGE } from '../consts/errorMessages';

import GIcon from '../components/CustomSVGs/GIcon';
import BrandLogo from '../components/CustomSVGs/BrandLogo';

function Login() {
  const [error, setError] = useState<string | null>(null);
  const location = useLocation();

  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const errorParam = searchParams.get('error') as ErrorType | null;
    if (errorParam) {
      setError(ERROR_MESSAGES[errorParam] || DEFAULT_ERROR_MESSAGE);
    }
  }, [location]);

  const handleLogin = () => {
    window.location.href = `${import.meta.env.VITE_API_ENDPOINT}/auth/google/signin`;
  };

  console.log('test url:', `${import.meta.env.VITE_API_ENDPOINT}/auth/google/signin`);
  return (

    <div className="bk_login grid lg:grid-cols-2 h-screen">
      {/* ------  Login  ------ */}
      <div className="bk_login__cta bg-black flex flex-col items-center justify-center px-4 py-6 sm:px-0 lg:py-0">
        <div className="max-w-md xl:max-w-xl">
          <h2 className="text-xl font-bold mb-8">Log in to your Q-KO Account</h2>
          {error && (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">
              <span className="block sm:inline">{error}</span>
            </div>
          )}
          <form className="space-y-4 md:space-y-6 w-full max-w-md xl:max-w-xl">
            <button
              className="text-white justify-center w-full border border-gray-700 hover:bg-gray-700 focus:ring-4 focus:outline-none focus:ring-[#4285F4]/50 font-medium rounded-lg text-sm px-5 py-2.5 text-center inline-flex items-center dark:focus:ring-[#4285F4]/55 me-2 mb-2  transition duration-500 ease-in-out hover:border-vivid-blue dark:hover:border-vivid-blue"
              onClick={handleLogin}
              name="Login with Google"
              type="button"
            >
              <GIcon />
              Continue with Google
              </button>
          </form>
        </div>
      </div>

      {/* ------  Copy  ------ */}
      <div className="bk_login__mktg flex items-center justify-center bg-white-smoke text-left p-6 sm:px-0 lg:py-0">
        <div className="max-w-2xl">
          <div className="flex flex-row">
            <BrandLogo className="h-10 w-10 me-2" />
            <a className="font-domus text-charcoal flex items-center text-4xl font-bold leading-none mb-4 hover:text-strong-violet" href="/">Q-KO</a>
          </div>
          <h1 className="font-domus text-4xl font-extrabold mb-4 text-charcoal">All of your books, in one place</h1>
          <p className="font-light leading-6 text-charcoal mb-4 lg:mb-8">By continuing, you acknowledge that you understand and agree to the <a className="" href="/terms-and-privacy">Terms & Conditions</a> and <a className="" href="/terms-and-privacy">Privacy Policy</a></p>
        </div>
      </div>
    </div>
  )
}

export default Login;
