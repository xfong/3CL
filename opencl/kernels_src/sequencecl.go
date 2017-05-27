package kernels_src

// Use the following lists to sequence order of file
// loads in order to build OpenCL Program
var OCLHeadersList = []string{
	"constants",
	"stdint",
	"stencil",
	"float3",
	"exchange",
	"atomicf",
	"reduce",
	"amul",
	"RNGmrg32k3a",
	"RNGmtgp",
	"sum"}

var OCLKernelsList = []string{
	"copypadmul2",
	"copyunpad",
	"crop",
	"addcubicanisotropy2",
	"pointwise_div",
	"divide",
	"adddmi",
	"adddmibulk",
	"dotproduct",
	"addexchange",
	"exchangedecode",
	"kernmulC",
	"kernmulRSymm2Dxy",
	"kernmulRSymm2Dz",
	"kernmulRSymm3D",
	"llnoprecess",
	"lltorque2",
	"madd2",
	"madd3",
	"setmaxangle",
	"minimize",
	"mul",
	"normalize2",
	"addoommfslonczewskitorque",
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
	"shiftbytesy",
	"shiftx",
	"shifty",
	"shiftz",
	"addslonczewskitorque2",
	"settemperature2",
	"settopologicalcharge",
	"adduniaxialanisotropy2",
	"vecnorm",
	"zeromask",
	"addzhanglitorque2",
	"mrg32k3a",
	"mtgp32_init_seed_kernel",
	"mtgp32_normal",
	"mtgp32_uniform",
	"square"}
