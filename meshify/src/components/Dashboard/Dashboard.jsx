import React, {useState} from 'react';
import logo from '../../logo.svg'; 
import meshes from '../../images/Meshes.svg';
import performance from '../../images/performance.svg';
import serviceHealth from '../../images/serviceHealth.svg';
import traffic from '../../images/traffic.svg';
import security from '../../images/security.svg';
import observability from '../../images/observability.svg';
import profile from '../../images/Profile.svg';
import notif from '../../images/notif.svg';

const Dashboard = () => {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [selectedContent, setSelectedContent] = useState(null);
  const [isServiceMeshesOpen, setIsServiceMeshesOpen] = useState(false);
  const [isPerformanceOpen, setIsPerformanceOpen] = useState(false);

  const toggleDropdown = () => {
    setIsDropdownOpen(!isDropdownOpen);
  };

  const toggleServiceMeshes = () => {
    setIsServiceMeshesOpen(!isServiceMeshesOpen);
    setIsPerformanceOpen(false);
    setSelectedContent('Service Meshes');
  }

  const togglePerformace = () => {
    setIsPerformanceOpen(!isPerformanceOpen);
    setIsServiceMeshesOpen(false);
    setSelectedContent('Performance');
  }

  return (
    <div className="fonts flex flex-col h-screen bg-gray-100">
      {/* Navbar */}
      <div className="bg-slate-200 text-white p-4 flex items-center">
        <img className="h-8 w-8 mr-2" src={logo} alt="Company Logo" />
        <div>
          <h2 className="company text-2xl text-indigo-700">Meshify</h2>

          
        </div>
        {setSelectedContent && (
            <span className='text-sm text-indigo-700 ml-2'>{selectedContent}</span>
          )}
        {/* Profile Section in Navbar */}
        <div className="ml-auto">
          <div className="relative">
            <button
              className="flex items-center focus:outline-none"
              onClick={toggleDropdown}
            >
              <span className="mr-2"><img src={notif} /></span>
              <span className=""><img src={profile} /></span>
              <svg
                className={`h-4 w-4 transition-transform transform ${
                  isDropdownOpen ? 'rotate-180' : ''
                }`}
                fill="none"
                stroke="black"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 9l-7 7-7-7"
                ></path>
              </svg>
            </button>
            {/* Dropdown Content */}
            {isDropdownOpen && (
              <div className="absolute right-0 mt-2 text-indigo-600 text-sm rounded-md shadow-lg">
                <button className="block w-full text-left px-4 py-2 text-gray-800 hover:bg-gray-200">
                  Profile
                </button>
                <button className="block w-full text-left px-4 py-2 text-gray-800 hover:bg-gray-200">
                  Get Token
                </button>
                <button className="block w-full text-left px-4 py-2 text-gray-800 hover:bg-gray-200">
                  Settings
                </button>
                <button className="block w-full text-left px-4 py-2 bg-gray-300 text-gray-800 hover:bg-gray-400">
                  Logout
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      

      {/* Sidebar */}
      <div className="w-64 bg-slate-200 text-white p-4 h-full">
        <h2 className="flex bg-indigo-700 p-1 items-center text-sm mb-2 "
        onClick={toggleServiceMeshes}>
          Service Meshes
          <img src={meshes} alt="Meshes" className="ml-auto" />
        </h2>
        
        {isServiceMeshesOpen && (
          <>
          <h2 className="flex bg-white text-indigo-700 mx-4 p-1 items-center text-sm mb-1">
          Istio
          <img src={meshes} alt="Meshes" className="ml-auto" />
        </h2>

        <h2 className="flex bg-white text-indigo-700 mx-4 p-1 items-center text-sm mb-1">
          Linkerd
          <img src={meshes} alt="Meshes" className="ml-auto" />
        </h2>

        <h2 className="flex bg-white text-indigo-700 mx-4 p-1 items-center text-sm mb-4">
          Cilium
          <img src={meshes} alt="Meshes" className="ml-auto" />
        </h2>
          </>
        )}

        <h2 className="flex bg-indigo-700 p-1 items-center text-sm mb-2"
        onClick={togglePerformace}>
          Performance
          <img src={performance} className="ml-auto" />
        </h2>

        {isPerformanceOpen && (
          <>
          <h2 className="flex bg-white text-indigo-700 mx-4 p-1 items-center text-sm mb-1">
          Prometheus
          <img src={meshes} alt="Meshes" className="ml-auto" />
        </h2>

        <h2 className="flex bg-white text-indigo-700 mx-4 p-1 items-center text-sm mb-4">
          Grafana
          <img src={meshes} alt="Meshes" className="ml-auto" />
        </h2>
          </>
        )}
        
        <h2 className="flex bg-indigo-700 p-1 items-center text-sm mb-4">
          Service health
          <img src={serviceHealth} className="ml-auto" />
        </h2>

        <h2 className="flex bg-indigo-700 p-1 items-center text-sm mb-4">
          Traffic Management
          <img src={traffic} className="ml-auto" />
        </h2>

        <h2 className="flex bg-indigo-700 p-1 items-center text-sm mb-4">
          Security
          <img src={security} className="ml-auto" />
        </h2>

        <h2 className="flex bg-indigo-700 p-1 items-center text-sm mb-4">
          Observability
          <img src={observability} className="ml-auto" />
        </h2>
        
      </div>

      <div className="bg-gray-300 text-black p-4 text-center">
        Built in Collaboration with
        <svg
  className="h-4 w-4 text-gray-500 inline-block mx-1"
  fill="currentColor"
  stroke="currentColor"
  viewBox="0 0 26 28"
  xmlns="http://www.w3.org/2000/svg"
>
  <path
    strokeLinecap="round"
    strokeLinejoin="round"
    strokeWidth="2"
    d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C16.09 3.81 17.76 3 19.5 3 22.58 3 25 5.42 25 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
  ></path>
</svg>

      </div>
    </div>
  );
};

export default Dashboard;