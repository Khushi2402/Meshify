import React from "react";
import { useState } from 'react';
import NavigationDrawer from "./Components/NavDrawer/NatvigationDrawer";
import { BrowserRouter, Routes, Route, Navigate} from "react-router-dom";
import Dashboard from "../src/pages/Dashboard"
import Settings from "./pages/Settings";
import Istio from "./pages/Istio";
import { disableBodyScroll } from 'body-scroll-lock';
import Auth from "./pages/Authentication/Auth";

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const handleLogin = () => {
    setIsAuthenticated(true);
  };
  return (
    <>
    <BrowserRouter>
    <Routes>
        <Route
          path="/"
          element={isAuthenticated ? <Dashboard /> : <Navigate to="/provider" />}
          />
        <Route
          path="/provider"
          element={<Auth onLogin={handleLogin} />}
           />
        <Route
          path="/dashboard"
          element={<Dashboard/>}
        />
        <Route path="/istio" element={<Istio/>} />
        <Route path="/settings" element={<Settings />} />
      </Routes>  
    </BrowserRouter>
    </>
  );
}

export default App;
