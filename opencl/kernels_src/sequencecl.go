package kernels_src

// Use the following lists to sequence order of file
// loads in order to build OpenCL Program
var OCLHeadersList = []string{
	"constants",
	"stdint",
	"clrng",
	"clprobdist",
	"stencil",
	"float3",
	"exchange",
	"atomicf",
	"reduce",
	"sum"	}

var OCLKernelsList = []string{
	"addcubicanisotropy",
	"crop",
	"adddmibulk",
	"adddmi",
	"divide",
	"dotproduct",
	"addexchange",
	"exchangedecode",
	"kernmulC",
	"kernmulRSymm2Dxy",
	"kernmulRSymm2Dz",
	"kernmulRSymm3D",
	"llnoprecess",
	"lltorque",
	"madd2",
	"madd3",
	"reducedot",
	"reducemaxabs",
	"reducemaxdiff",
	"reducemaxvecdiff2",
	"reducemaxvecnorm2",
	"reducesum",
	"regionaddv",
	"regiondecode",
	"regionselect",
	"resize",
	"shiftbytes",
	"shiftx",
	"addslonczewskitorque",
	"settemperature",
	"settopologicalcharge",
	"adduniaxialanisotropy",
	"vecnorm",
	"zeromask",
	"addzhanglitorque",
	"square"	}