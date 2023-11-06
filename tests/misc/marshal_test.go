package misc

import (
	"encoding/json"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/stretchr/testify/require"
	"testing"
)

// Round trip for blocks
func Test_CapellaSignedBlockRoundTripJSON(t *testing.T) {
	var signedBlockCapella capella.SignedBeaconBlock
	signedBlockCapellaJSON := []byte("{\"message\":{\"slot\":\"28\",\"proposer_index\":\"136\",\"parent_root\":\"0x004f5d49557176647e0337bdc45d6d7c88bb0ed4287511a5169c65c693f3544b\",\"state_root\":\"0xc3f067810d2a6c195e5e0b25cf05b32763dbb9d423c0e88738ef8b4849d7acb4\",\"body\":{\"randao_reveal\":\"0x8a70bbb7873cf982213cc2d4886a601a173a103a0767a0013a1eb7239ea88fac7dbcc049adcae77d25186017ec9a1b8e17996aa58780f398c157b5d2bc0b5118cd5a75c57711bb5229ce4448e5170fe7ecafbd13ab5e5f863f1528c5fb6ea928\",\"eth1_data\":{\"deposit_root\":\"0xd70a234731285c6804c2a4f56711ddb8c82c99740f207854891028af34e27e5e\",\"deposit_count\":\"0\",\"block_hash\":\"0x32f9f471274b1a60423dc3ae47152ea9508b60e1adb4ef712e010d089d561835\"},\"graffiti\":\"0x6c6f6465737461722d676574682d300000000000000000000000000000000000\",\"proposer_slashings\":[],\"attester_slashings\":[],\"attestations\":[{\"aggregation_bits\":\"0xeb01\",\"data\":{\"slot\":\"27\",\"index\":\"3\",\"beacon_block_root\":\"0x004f5d49557176647e0337bdc45d6d7c88bb0ed4287511a5169c65c693f3544b\",\"source\":{\"epoch\":\"1\",\"root\":\"0x111a272d76507bf1bba46fbf4d6f6b0824a2fc8c40b94f9a4aaea1ed129e8362\"},\"target\":{\"epoch\":\"3\",\"root\":\"0x33076cada6f3e90d2921821419f5f7a8e9cf43c02e48987a6b1ac9e89aa0d7a8\"}},\"signature\":\"0xa6b968124685a3f60141ed07562f324a48ec93523bc4d900ee4435abc449561e095af0c3758fe6c41a881f759239efd305961f3f44fd41d82f905e11f32dad1e05d41d1f7efb3a9238e2a2e4ce48903fe07e4c5167e149b919755a332d88a600\"},{\"aggregation_bits\":\"0xeb\",\"data\":{\"slot\":\"27\",\"index\":\"0\",\"beacon_block_root\":\"0x004f5d49557176647e0337bdc45d6d7c88bb0ed4287511a5169c65c693f3544b\",\"source\":{\"epoch\":\"1\",\"root\":\"0x111a272d76507bf1bba46fbf4d6f6b0824a2fc8c40b94f9a4aaea1ed129e8362\"},\"target\":{\"epoch\":\"3\",\"root\":\"0x33076cada6f3e90d2921821419f5f7a8e9cf43c02e48987a6b1ac9e89aa0d7a8\"}},\"signature\":\"0xb0c4fc2f49f818925245400aae1a15927a250ef1724ca50a4488570966af7e11b1827bb5f67e283feaf3a2caa043c3a60bd8f05e7a247d89666967f81479f11f6712fe03e88d889d3ebc68231a0dd3fdf82c7da9174b21369898d265eb8d636c\"},{\"aggregation_bits\":\"0xeb\",\"data\":{\"slot\":\"27\",\"index\":\"2\",\"beacon_block_root\":\"0x004f5d49557176647e0337bdc45d6d7c88bb0ed4287511a5169c65c693f3544b\",\"source\":{\"epoch\":\"1\",\"root\":\"0x111a272d76507bf1bba46fbf4d6f6b0824a2fc8c40b94f9a4aaea1ed129e8362\"},\"target\":{\"epoch\":\"3\",\"root\":\"0x33076cada6f3e90d2921821419f5f7a8e9cf43c02e48987a6b1ac9e89aa0d7a8\"}},\"signature\":\"0xa48c03b08adab5535677b6fee6077c685353a0026e2210160059991c482dce786b96b73211057c475ac00b25586fc6c70e2438e45db7c1e69e4c52d874d45ff7c07132e4e9cab9ee0c1f2ffd49c02aadc37e016c2fad49b11d83ff111d8868d9\"},{\"aggregation_bits\":\"0xc901\",\"data\":{\"slot\":\"27\",\"index\":\"1\",\"beacon_block_root\":\"0x004f5d49557176647e0337bdc45d6d7c88bb0ed4287511a5169c65c693f3544b\",\"source\":{\"epoch\":\"1\",\"root\":\"0x111a272d76507bf1bba46fbf4d6f6b0824a2fc8c40b94f9a4aaea1ed129e8362\"},\"target\":{\"epoch\":\"3\",\"root\":\"0x33076cada6f3e90d2921821419f5f7a8e9cf43c02e48987a6b1ac9e89aa0d7a8\"}},\"signature\":\"0x98c01067412c211087f9be9feb5182b8dd94cc39c753b167aa4e1f189a4a7eec1d4e85fbbc3a8eb91505050896e563ff155bee363dc889e31d8797e7de253a90fe859f8d7d1f5aa8b3b2bd8c1d3437a9907ffd3e859ed52b2e0178015b322be2\"},{\"aggregation_bits\":\"0xfe\",\"data\":{\"slot\":\"20\",\"index\":\"2\",\"beacon_block_root\":\"0xac245bed6119365514b0cb012719a38940cd808eb2df1e44b1630c6d751f37de\",\"source\":{\"epoch\":\"0\",\"root\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"},\"target\":{\"epoch\":\"2\",\"root\":\"0x0e8e1ec8dbc201118297a63078d3ec85f3da8965f2fb28d69657aabc418eb283\"}},\"signature\":\"0xae15a3355891318210a243ff843f54226d2f20d5b6c00fc902099405b728c0d589c74f9356cf8319a09c46c5c542f73f0c8b7f4aa6a2f8698027849bc0cb1e31e3283ecf5514ff9cb700ff88116495afd6a28f605bc6692ef64c02f305742a2a\"},{\"aggregation_bits\":\"0xaf\",\"data\":{\"slot\":\"20\",\"index\":\"0\",\"beacon_block_root\":\"0xac245bed6119365514b0cb012719a38940cd808eb2df1e44b1630c6d751f37de\",\"source\":{\"epoch\":\"0\",\"root\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"},\"target\":{\"epoch\":\"2\",\"root\":\"0x0e8e1ec8dbc201118297a63078d3ec85f3da8965f2fb28d69657aabc418eb283\"}},\"signature\":\"0xa05af337da57b8b69b8a531afde2e5c7c8328dbf4a506ab27a4a5d97ba33ccde43c23094b1b170ca3b9061eadc3419090156155abdfff6bf47ae42cbb2fc2dafaed837e81132c3c97d04543338f934a28d55615610f3285d459a272ad4df4522\"},{\"aggregation_bits\":\"0x4e01\",\"data\":{\"slot\":\"20\",\"index\":\"3\",\"beacon_block_root\":\"0xac245bed6119365514b0cb012719a38940cd808eb2df1e44b1630c6d751f37de\",\"source\":{\"epoch\":\"0\",\"root\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"},\"target\":{\"epoch\":\"2\",\"root\":\"0x0e8e1ec8dbc201118297a63078d3ec85f3da8965f2fb28d69657aabc418eb283\"}},\"signature\":\"0x83117b59c1b3f7004cf7adcb7267c08cd8762d0d935edc7541ebb68ea95cc12783842710152799c654aebf161141a81b0b65c9bf16c18467bc3ece9f843c1f392a87d0aaefa811f47c1ed6cde9e3e974bcda30c75bd856b4fc64ced006d4863b\"},{\"aggregation_bits\":\"0xc201\",\"data\":{\"slot\":\"20\",\"index\":\"1\",\"beacon_block_root\":\"0xac245bed6119365514b0cb012719a38940cd808eb2df1e44b1630c6d751f37de\",\"source\":{\"epoch\":\"0\",\"root\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"},\"target\":{\"epoch\":\"2\",\"root\":\"0x0e8e1ec8dbc201118297a63078d3ec85f3da8965f2fb28d69657aabc418eb283\"}},\"signature\":\"0x93f16988d7c952d7e8f7635dc9004cfebca4be4c857401ad57959491565251196b1cf642ebbb95bdf222d43688e0c27409bf21de417dafadfcc9be681971f5f56ebd51903ef0085e7b862560dedd63c7accc68d4722f11cc7d43bb9339f33902\"}],\"deposits\":[],\"voluntary_exits\":[],\"sync_aggregate\":{\"sync_committee_bits\":\"0x96b9ae6d\",\"sync_committee_signature\":\"0xa35cf81718d92baee9a1378feebd86a2f23ba58a0c9acb4e28aecfc8674fb3d5ad528781198e99fdadbf7f188e9d24b40e6db65e20bc43173f800b24a295e84c8e0ec7d1f6575eec25ff6cf83ed20c6e9fb295c832d24587e56ab7c827e3121f\"},\"execution_payload\":{\"parent_hash\":\"0x5d6b1827b60907d7c0867496369b5c9c5e08d41402bfdbf23e365328be20f336\",\"fee_recipient\":\"0x00000000219ab540356cbb839cbe05303d7705fa\",\"state_root\":\"0xfc98f963dbcf16dd42c1e25ef6c502cf45df8b47a0da97054fabf445e11eb647\",\"receipts_root\":\"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421\",\"logs_bloom\":\"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"prev_randao\":\"0x0b70cb7af6f71e1998f24514b64b5bfea9188fdb74b2e874a1dd55e67718a626\",\"block_number\":\"15\",\"gas_limit\":\"25368701\",\"gas_used\":\"0\",\"timestamp\":\"1699295226\",\"extra_data\":\"0xd883010d05846765746888676f312e32302e33856c696e7578\",\"base_fee_per_gas\":\"134933816\",\"block_hash\":\"0xef0fb2fd252088c2262383070da059bb4db17575f2942f24dc5c17b3cd8ab732\",\"transactions\":[],\"withdrawals\":[]},\"bls_to_execution_changes\":[]}},\"signature\":\"0xa66689f1e7878ffcd80bf646341776010d927b0cfb24c6441ff6bb11f870958642738642d980a6fc8443893de07ce94a160b0c8fae1c984116485d0ba9e8dc7c065acc41eb0ff35a17c024db5eaa88c440851417e2dee3a486e9229df8efe49f\"}")
	err := json.Unmarshal(signedBlockCapellaJSON, &signedBlockCapella)
	require.NoError(t, err)
	serializedBlock, err := json.Marshal(signedBlockCapella)
	require.Equal(t, signedBlockCapellaJSON, serializedBlock)
}
