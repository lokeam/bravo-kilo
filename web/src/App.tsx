import { Suspense } from 'react'
import { Route, Routes } from 'react-router-dom'
import Home from './pages/home'
import Login from './pages/login'
import Library from './pages/library'

import './App.css'

function App() {

  return (
    <Suspense fallback={<h1>Loading...</h1>}>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/library" element={<Library />} />
        <Route path="/login" element={<Login />}/>
      </Routes>
    </Suspense>
  )
}

export default App
