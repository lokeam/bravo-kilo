import GIcon from '../components/CustomSVGs/GIcon';

function Login() {
  const handleLogin = () => {
    window.location.href = `${import.meta.env.VITE_API_ENDPOINT}/auth/google/signin`;
  };

  console.log('test url:', `${import.meta.env.VITE_API_ENDPOINT}/auth/google/signin`);
  return (

    <div className="bk_login grid lg:grid-cols-2 h-screen">
      {/* ------  Login  ------ */}
      <div className="bk_login__cta bg-black flex flex-col items-center justify-center px-4 py-6 sm:px-0 lg:py-0">
        <div className="max-w-md xl:max-w-xl">
          <h2 className="text-xl font-bold mb-8">Log in to your Bravo Kilo Account</h2>
          <form className="space-y-4 md:space-y-6 w-full max-w-md xl:max-w-xl">
            <button
              className="text-white justify-center w-full border border-gray-700 hover:bg-gray-700 focus:ring-4 focus:outline-none focus:ring-[#4285F4]/50 font-medium rounded-lg text-sm px-5 py-2.5 text-center inline-flex items-center dark:focus:ring-[#4285F4]/55 me-2 mb-2"
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
      <div className="bk_login__mktg flex items-center justify-center bg-majorelle text-left px-4 py-6 sm:px-0 lg:py-0">
        <div className="max-w-md">
          <a className="flex items-center text-2xl text-white font-semibold leading-none mb-4" href="/">Bravo Kilo</a>
          <h1 className="text-4xl font-extrabold mb-4">Catchline here</h1>
          <p className="text-az-white font-light leading-6 mb-4 lg:mb-8">Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.</p>
        </div>
      </div>
    </div>
  )
}

export default Login;
