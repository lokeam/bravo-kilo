import { Suspense } from 'react';
import { Route, Routes } from 'react-router-dom';
import { AuthProvider } from './components/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Home from './pages/home';
import Login from './pages/login';
import Library from './pages/library';
import BookDetail from './pages/bookDetail';

import NotFound from './pages/notFound';

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
          <Route path="/library/books/:bookID" element={
            <ProtectedRoute>
              <BookDetail />
            </ProtectedRoute>
          } />
          <Route path="*" element={<NotFound />}/>
        </Routes>
      </Suspense>
    </AuthProvider>
  );
}

export default App;
