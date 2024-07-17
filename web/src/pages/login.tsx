
const Login = () => {
  const handleLogin = () => {
    window.location.href = `${import.meta.env.VITE_API_ENDPOINT}/google-signin`;
  };


  console.log('test url:', `${import.meta.env.VITE_API_ENDPOINT}/google-signin`);
  return (
    <div className="bk_login">
      <h1>Login</h1>
        <form>
          <button type="button" onClick={handleLogin}>Sign in with Google</button>
        </form>
    </div>
  )
}

export default Login;
