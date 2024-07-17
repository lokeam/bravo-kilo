import { Suspense } from 'react'
import { Route, Routes } from 'react-router-dom'
import { AuthProvider } from './components/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Home from './pages/home'
import Login from './pages/login'
import Library from './pages/library'

import './App.css'


function App() {
  return (
    <AuthProvider>
      <Suspense fallback={<h1>Loading...</h1>}>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />}/>
          <Route path="/library" element={
            <ProtectedRoute>
              <Library />
            </ProtectedRoute>
          } />
        </Routes>
      </Suspense>
    </AuthProvider>
  );
}

export default App;
